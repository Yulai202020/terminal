#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int main() {
    if (isatty(0)) {
        printf("Terminal!\n");
    } else {
        printf("No terminal!\n");
    }
}