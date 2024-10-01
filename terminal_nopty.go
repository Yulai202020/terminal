package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/chzyer/readline"
)

const LogFile = "log_go.txt"

func runCommand(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return outBuf.String(), errBuf.String(), err
		}
	case <-time.After(5 * time.Second):
		cmd.Process.Kill()
		return outBuf.String(), errBuf.String(), fmt.Errorf("command timed out")
	}

	return outBuf.String(), errBuf.String(), nil
}

func getpwd() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func chdir(newDir string) error {
	err := os.Chdir(newDir)
	if err != nil {
		return fmt.Errorf("Error on chdir: %w", err)
	}
	return nil
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func main() {
	cwd, err := getpwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer file.Close()

	for {
		rl, err := readline.New(cwd + "> ")
		if err != nil {
			fmt.Println("Error creating readline instance:", err)
			return
		}
		defer rl.Close()

		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				continue
			}
			fmt.Println("Error reading line:", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		command := parts[0]
		args := parts[1:]

		if line == "exit" {
			break
		} else if command == "cd" {
			err := chdir(strings.Join(args, " "))
			if err != nil {
				fmt.Printf("cd error: %s\n", err)
			}
			continue
		}

		if !commandExists(command) {
			fmt.Println("Command doesn't exist!")
			continue
		}

		_, err = file.WriteString("Command: " + line + "\n")
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
		}

		output, errOutput, err := runCommand(command, args...)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
		}

		_, err = file.WriteString(output + "===================\n")
		if err != nil {
			fmt.Printf("Error writing to log file: %v\n", err)
		}

		fmt.Print(output)
		if errOutput != "" {
			fmt.Print(errOutput)
		}
	}
}
