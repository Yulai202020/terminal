pty:
	gcc terminal.c -o terminal -lreadline -lutil -lstring -lm
nopty:
	gcc terminal_nopty.c -o terminal -lreadline -lstring -lm
