package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "strings"
    "github.com/chzyer/readline"
)

const LogFile = "log.txt"
var IgnoreList = []string{"vim", "nano"}

func isInIgnoreList(item string) bool {
    for _, ignoreItem := range IgnoreList {
        if ignoreItem == item {
            return true
        }
    }
    return false
}

func logCommand(command string) error {
    file, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return fmt.Errorf("error opening log file: %w", err)
    }
    defer file.Close()

    _, err = file.WriteString(command + "\n")
    if err != nil {
        return fmt.Errorf("error writing to log file: %w", err)
    }
    return nil
}

func runCommand(command string, args ...string) error {
    logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println("Error opening log file:", err)
        return err
    }
    defer logFile.Close()

    var multiWriter io.Writer
    if isInIgnoreList(command) {
        multiWriter = io.MultiWriter(os.Stdout)
    } else {
        multiWriter = io.MultiWriter(os.Stdout, logFile)
    }

    cmd := exec.Command(command, args...)

    // Set the output to the current process's stdout
    cmd.Stdout = multiWriter
    cmd.Stderr = os.Stderr
    cmd.Stdin = os.Stdin

    // Start the command
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("error starting command: %w", err)
    }

    return cmd.Wait()
}


func getpwd() (string, error) {
    return os.Getwd()
}

func chdir(newDir string) error {
    return os.Chdir(newDir)
}

func commandExists(command string) bool {
    _, err := exec.LookPath(command)
    return err == nil
}

func main() {
    for {
        cwd, err := getpwd()
        if err != nil {
            fmt.Println("Error getting current directory:", err)
            continue
        }

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

        if command == "exit" {
            break
        } else if command == "cd" {
            if err := chdir(strings.Join(args, " ")); err != nil {
                fmt.Println("cd error:", err)
            }
            continue
        }

        if !commandExists(command) {
            fmt.Println("Command doesn't exist!")
            continue
        }

        if err := logCommand("Command: " + line); err != nil {
            fmt.Println(err)
            continue
        }

        if err := runCommand(command, args...); err != nil {
            fmt.Println("Command finished with error:", err)
        }
        
        logCommand("===================")
    }
}
