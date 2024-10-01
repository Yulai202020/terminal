package main

import (
	"bytes"
	"fmt"
	"io"
    "os"
	"os/exec"
	"time"
    "github.com/chzyer/readline"
	"strings"

	"github.com/creack/pty"
)

const LogFile = "log_go.txt"

func runCommand(command string, args ...string) (string, string, error) {
	cmd := exec.Command(command, args...)

	var outBuf, errBuf bytes.Buffer

	pty, err := pty.Start(cmd)
	if err != nil {
		return "", "", err
	}
	defer pty.Close()

	go func() {
		io.Copy(&outBuf, pty)
	}()

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
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

    return cwd, err
}

func chdir(newDir string) (error) {
    err := os.Chdir(newDir)
    if err != nil {
        fmt.Println("Error on chdir:", err);
    }
    return err
}

func commandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

func main() {
    for {
        // init readline
        cwd, err := getpwd();
        rl, err := readline.New(cwd + "> ");

    
        if err != nil {
            fmt.Println("Error creating readline instance:", err)
            return;
        }
    
        defer rl.Close();

        // get input
		line, err := rl.Readline();

		if err != nil {
			if err == readline.ErrInterrupt {
                continue;
			}

			fmt.Println("Error reading line:", err)
			return
		}

        // check is variable line is not empty
        line = strings.TrimSpace(line) // Remove leading and trailing whitespace
		if line == "" {
			continue
		}

        // split command
        parts := strings.Split(line, " ")
        command := parts[0];
        args := strings.Join(parts[1:], " ");

        // if exit, exit from program
		if line == "exit" {
			break
		} else if parts[0] == "cd" {
            err := chdir(args);

            if (err != nil) {
                fmt.Printf("cd error: %s\n", err)
                break;
            }

            continue
        }

        if (!commandExists(command)) {
            fmt.Println("Command doesnt exist!");
            continue
        }

        // open file (log anythink)
        file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println("Error opening log file:", err)
            return
        }
        defer file.Close() 

        // write command
        _, err = file.WriteString("Command: "+line+"\n");

		if err != nil {
			fmt.Printf("Error on writing to log file: %v\n", err);
		}

        // get output of command
		output, errOutput, err := runCommand(command, parts[1:]...);

        // write to file output
        _, err = file.WriteString(output+"===================\n")

		if err != nil {
			fmt.Printf("Error on writing to log file: %v\n", err);
		}

        // finally print output with stderr
        fmt.Print(output);
        fmt.Print(errOutput);
	}
}