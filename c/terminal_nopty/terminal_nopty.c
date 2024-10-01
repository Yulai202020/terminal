#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include <unistd.h>

#include <pty.h>
#include <fcntl.h>
#include <termios.h>

#include <sys/wait.h>
#include <sys/types.h>
#include <sys/select.h>

#include <readline/readline.h>
#include <readline/history.h>

#include <yulai/string.h>

#define LOG_FILE "log.txt"

char* replace_envs(const char* original) {
    size_t size = strlen(original);
    size_t new_size = 0;
    
    char* copy = malloc(1);
    
    if (!copy) {
        return NULL;
    }
    
    copy[0] = '\0';

    for (size_t i = 0; i < size; i++) {
        if (original[i] == '$') {
            i++;
            size_t start = i;

            while (isalnum(original[i]) || original[i] == '_') {
                i++;
            }

            size_t end = i;

            char* env_key = substring(original, start, end);
            char* env_value = getenv(env_key);

            if (env_value) {
                size_t env_value_len = strlen(env_value);
                copy = stringSumRealloc(copy, env_value);

                new_size += env_value_len;
            } else {
                size_t var_len = end - start + 1;
                copy = stringSumRealloc(copy, stringSum("$", env_key));

                new_size += var_len;
            }
            free(env_key);
            i--;
        } else {
            copy = realloc(copy, new_size + 2);

            if (!copy) {
                return NULL;
            }

            copy[new_size++] = original[i];
            copy[new_size] = '\0';
        }
    }

    return copy;
}

char* str_replace(const char* original, const char* old, const char* new) {
    int count = 0;
    const char* tmp = original;
    while ((tmp = strstr(tmp, old))) {
        count++;
        tmp += strlen(old);
    }

    size_t new_length = strlen(original) + count * (strlen(new) - strlen(old)) + 1;
    char* result = malloc(new_length);
    if (!result) return NULL;

    char* ptr = result;
    while (*original) {
        if (strstr(original, old) == original) {
            strcpy(ptr, new);
            ptr += strlen(new);
            original += strlen(old);
        } else {
            *ptr++ = *original++;
        }
    }

    *ptr = '\0';
    return result;
}

char* get_current_working_directory() {
    char* cwd = malloc(1024);
    if (cwd == NULL) {
        perror("malloc failed");
        return NULL;
    }

    if (getcwd(cwd, 1024) == NULL) {
        perror("getcwd() error");
        free(cwd);
        return NULL;
    }

    return cwd;
}

void run_as_pty(const char* command) {
    int master_fd, slave_fd;
    pid_t pid;
    struct termios orig_termios;
    int fds[2];

    if (pipe(fds) == -1) {
        perror("openpty");
        exit(EXIT_FAILURE);
    }

    master_fd = fds[0];
    slave_fd = fds[1];

    pid = fork();
    if (pid == -1) {
        perror("fork");
        exit(EXIT_FAILURE);
    }

    if (pid == 0) {
        close(master_fd);
        
        setsid();
        dup2(slave_fd, STDIN_FILENO);
        dup2(slave_fd, STDOUT_FILENO);
        dup2(slave_fd, STDERR_FILENO);
        close(slave_fd);

        char* cmd = strdup(command);
        char* argv[64];
        int i = 0;

        char* token = strtok(cmd, " ");
        while (token != NULL && i < 63) {
            argv[i++] = token;
            token = strtok(NULL, " ");
        }
        argv[i] = NULL;

        int rc = execvp(argv[0], argv);

        if (rc == -1) {
            printf("Command not found!\n");
        }

        exit(EXIT_FAILURE);
    } else {
        close(slave_fd);

        tcgetattr(STDIN_FILENO, &orig_termios);
        struct termios new_termios = orig_termios;
        new_termios.c_lflag &= ~(ICANON | ECHO);
        tcsetattr(STDIN_FILENO, TCSANOW, &new_termios);

        char buffer[256];
        fd_set readfds;

        FILE* ptr = fopen(LOG_FILE, "a");
        fprintf(ptr, "Ouput:\n");

        while (1) {
            FD_ZERO(&readfds);
            FD_SET(master_fd, &readfds);
            FD_SET(STDIN_FILENO, &readfds);

            int max_fd = master_fd > STDIN_FILENO ? master_fd : STDIN_FILENO;
            int activity = select(max_fd + 1, &readfds, NULL, NULL, NULL);

            if (activity < 0) {
                perror("select");
                break;
            }

            if (FD_ISSET(master_fd, &readfds)) {
                int bytes_read = read(master_fd, buffer, sizeof(buffer));

                if (bytes_read > 0) {
                    write(STDOUT_FILENO, buffer, bytes_read);
                    fprintf(ptr, "%s", buffer);
                } else {
                    break;
                }
            }

            if (FD_ISSET(STDIN_FILENO, &readfds)) {
                int bytes_read = read(STDIN_FILENO, buffer, sizeof(buffer));

                if (bytes_read > 0) {
                    if (strncmp(buffer, "exit", 4) == 0) {
                        break;
                    }

                    write(master_fd, buffer, bytes_read);
                } else {
                    break;
                }
            }
        }

        tcsetattr(STDIN_FILENO, TCSANOW, &orig_termios);
        waitpid(pid, NULL, 0);

        fclose(ptr);
    }
}

int main() {
    char* command;

    printf("Simple Command Terminal\n");
    printf("Type a command and press Enter (type 'exit' or ctrl+C to quit):\n");

    while (1) {
        char prompt[1025];
        snprintf(prompt, sizeof(prompt), "%s> ", get_current_working_directory());

        command = readline(prompt);

        if (command == NULL) {
            break;
        }

        size_t len = strlen(command);
        while (len > 0 && isspace((unsigned char)command[len - 1])) {
            command[--len] = '\0';
        }

        if (len == 0) {
            free(command);
            continue;
        }

        command = str_replace(command, "~", getenv("HOME"));
        command = replace_envs(command);

        if (strcmp(command, "exit") == 0) {
            free(command);
            break;
        } else if (strcmp(command, "clear") == 0) {
            system("clear");
            free(command);
            continue;
        } else if (strncmp(command, "cd ", 3) == 0) {
            char* path = command + 3;

            if (chdir(path) != 0) {
                perror("cd failed");
            }

            free(command);
            continue;
        }

        FILE* ptr = fopen(LOG_FILE, "a");
        fprintf(ptr, "Command: %s\n", command);
        fclose(ptr);

        run_as_pty(command);

        ptr = fopen(LOG_FILE, "a");
        fprintf(ptr, "=================\n", command);
        fclose(ptr);

        add_history(command);

        free(command);
    }

    return 0;
}
