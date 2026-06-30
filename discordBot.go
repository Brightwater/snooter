package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/phuslu/log"
	"github.com/pkg/errors"
)

// DiscordBot handles Discord WebSocket connection and message operations.
type DiscordBot struct {
	session   *discordgo.Session
	channelID string
}

// NewDiscordBot initializes a new DiscordBot instance.
func NewDiscordBot(token, channelID string) (*DiscordBot, error) {
	if token == "" {
		return nil, errors.New("discord token is required")
	}
	if channelID == "" {
		return nil, errors.New("discord channel ID is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create discordgo session")
	}

	bot := &DiscordBot{
		session:   session,
		channelID: channelID,
	}

	session.AddHandler(bot.onReady)

	return bot, nil
}

// Start opens the WebSocket connection to Discord.
func (b *DiscordBot) Start() error {
	log.Info().Msg("Connecting to Discord...")
	if err := b.session.Open(); err != nil {
		return errors.Wrap(err, "failed to open discord connection")
	}
	return nil
}

// Stop closes the Discord connection.
func (b *DiscordBot) Stop() error {
	log.Info().Msg("Disconnecting from Discord...")
	return b.session.Close()
}

// RegisterHandler registers a generic event handler to the discordgo session.
func (b *DiscordBot) RegisterHandler(handler interface{}) {
	b.session.AddHandler(handler)
}

func (b *DiscordBot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Info().Str("username", s.State.User.Username).Msg("Discord bot is ready and connected!")
}

// SendNotification sends an update report embed with interactive action buttons to the target channel.
// It returns the sent message's ID.
func (b *DiscordBot) SendNotification(report *UpdateReport) (string, error) {
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🚀 Update Available: %s", report.AppName),
		Description: fmt.Sprintf("**Current Version:** `%s` ➡️ **Target Version:** `%s`\n**Update Method:** %s", report.CurrentVersion, report.NewVersion, report.UpdateMethod),
		Color:       0x3498db, // Default Blue
	}

	if report.ReleaseURL != "" {
		embed.Description += fmt.Sprintf("\n**Release Notes:** [View on GitHub](%s)", report.ReleaseURL)
	}

	if report.Analysis != nil {
		var riskColor int
		riskEmoji := "🟢"
		lowerRisk := strings.ToLower(report.Analysis.AutoUpdateRisk)
		if strings.Contains(lowerRisk, "caution") {
			riskEmoji = "🟡"
			riskColor = 0xf1c40f // Yellow
		} else if strings.Contains(lowerRisk, "high") || strings.Contains(lowerRisk, "critical") {
			riskEmoji = "🔴"
			riskColor = 0xe74c3c // Red
		} else {
			riskColor = 0x2ecc71 // Green
		}
		embed.Color = riskColor

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   riskEmoji + " Risk Level",
			Value:  report.Analysis.AutoUpdateRisk,
			Inline: true,
		})

		if report.Analysis.ChangeSummary != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "📝 AI Summary",
				Value:  report.Analysis.ChangeSummary,
				Inline: false,
			})
		}

		if report.Analysis.SecurityReport != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "🛡️ Security Report",
				Value:  report.Analysis.SecurityReport,
				Inline: false,
			})
		}
	}

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Proceed",
					Style:    discordgo.SuccessButton,
					CustomID: fmt.Sprintf("snooter:proceed:%s:%s", report.AppName, report.NewVersion),
					Emoji: &discordgo.ComponentEmoji{
						Name: "🚀",
					},
				},
				discordgo.Button{
					Label:    "Snooze Version",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("snooter:snooze_version:%s:%s", report.AppName, report.NewVersion),
					Emoji: &discordgo.ComponentEmoji{
						Name: "💤",
					},
				},
				discordgo.Button{
					Label:    "Snooze Report",
					Style:    discordgo.SecondaryButton,
					CustomID: fmt.Sprintf("snooter:snooze_report:%s:%s", report.AppName, report.NewVersion),
					Emoji: &discordgo.ComponentEmoji{
						Name: "⏰",
					},
				},
			},
		},
	}

	msg, err := b.session.ChannelMessageSendComplex(b.channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to send channel message")
	}

	return msg.ID, nil
}

// RespondDeferred acknowledges a component interaction deferred-ly to keep it open.
func (b *DiscordBot) RespondDeferred(i *discordgo.Interaction) error {
	return errors.Wrap(
		b.session.InteractionRespond(i, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		}),
		"failed to send deferred interaction response",
	)
}

// SendFollowup sends a follow-up message to a deferred interaction.
func (b *DiscordBot) SendFollowup(i *discordgo.Interaction, content string) error {
	_, err := b.session.FollowupMessageCreate(i, true, &discordgo.WebhookParams{
		Content: content,
	})
	return errors.Wrap(err, "failed to send interaction follow-up message")
}

// DisableMessageButtons edits a message, disabling all buttons inside its action row components.
func (b *DiscordBot) DisableMessageButtons(msg *discordgo.Message) error {
	if msg == nil || len(msg.Components) == 0 {
		return nil
	}

	var updatedComponents []discordgo.MessageComponent
	for _, rawComp := range msg.Components {
		row, ok := rawComp.(*discordgo.ActionsRow)
		if !ok {
			continue
		}

		var updatedButtons []discordgo.MessageComponent
		for _, btnComp := range row.Components {
			btn, ok := btnComp.(*discordgo.Button)
			if !ok {
				continue
			}
			btn.Disabled = true
			updatedButtons = append(updatedButtons, *btn)
		}
		updatedComponents = append(updatedComponents, discordgo.ActionsRow{Components: updatedButtons})
	}

	_, err := b.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         msg.ID,
		Channel:    msg.ChannelID,
		Components: &updatedComponents,
	})
	return errors.Wrap(err, "failed to edit message components")
}
