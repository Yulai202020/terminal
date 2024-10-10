package main

import (
    "io"
    "os"
    "fmt"
    "sync"
    "os/exec"
	"strings"

    "golang.org/x/term"
    "github.com/creack/pty"
    "github.com/chzyer/readline"
)

const LogFile = "log.txt"
var IgnoreList = []string{"vim", "nano", "nvim", "vi"}

func isInIgnoreList(item string) bool {
    for _, ignoreItem := range IgnoreList {
        if ignoreItem == item {
            return true
        }
    }
    return false
}

func runCommand(command string, args ...string) {
    logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening log file:", err)
        return
    }
    defer logFile.Close()

    var multiWriter io.Writer
    if isInIgnoreList(command) {
        multiWriter = io.MultiWriter(os.Stdout)
    } else {
        multiWriter = io.MultiWriter(os.Stdout, logFile)
    }

    cmd := exec.Command(command, args...)
    ptmx, err := pty.Start(cmd)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error starting command: %v\n", err)
        return
    }
    defer ptmx.Close()

    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error setting terminal to raw mode: %v\n", err)
        return
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    var wg sync.WaitGroup
    wg.Add(1)

    go func() {
        defer wg.Done()
        _, _ = io.Copy(multiWriter, ptmx)
    }()

    // Copy from stdin to ptmx
    go func() {
        _, _ = io.Copy(ptmx, os.Stdin)
    }()

    // Wait for command to finish
    err = cmd.Wait()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Command finished with error: %v\n", err)
    }

    // Wait for the output goroutine to finish
    wg.Wait()
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
    // whi
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
        
        // write command
        _, err = file.WriteString("Command: "+line+"\n");

		if err != nil {
			fmt.Printf("Error on writing to log file: %v\n", err);
		}

        file.Close()

        // i close and open cuz runCommand also opens this file so its can break

        // get output of command
		runCommand(command, parts[1:]...);

        file, err = os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            fmt.Println("Error opening log file:", err)
            return
        }

        _, err = file.WriteString("===================\n")

		if err != nil {
			fmt.Printf("Error on writing to log file: %v\n", err);
		}

        file.Close();
	}
}
