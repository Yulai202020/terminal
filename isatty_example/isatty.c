#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int main() {
    if (isatty(0) && isatty(1) && isatty(2)) {
        printf("Terminal!\n");
    } else {
        printf("No terminal!\n");
    }
}
