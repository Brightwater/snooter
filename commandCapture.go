package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

func RunCommandCaptureOutput(cmd *exec.Cmd) (string, error) {
	// cmd := exec.Command(command, args...)

	// Get stdout pipe
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	// Get stderr pipe
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)

	// Function to copy output both to stdout and buffer

	go copyOutput(stdoutPipe, &wg, &buf)
	go copyOutput(stderrPipe, &wg, &buf)

	wg.Wait()

	err = cmd.Wait()

	return buf.String(), err
}

func copyOutput(r io.Reader, wg *sync.WaitGroup, buf *bytes.Buffer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)            // print to stdout in real time
		buf.WriteString(line + "\n") // save to buffer
	}
	wg.Done()
}
