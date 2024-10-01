#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>

int main() {
    if (isatty(0)) {
        printf("TERMINAL!\n");
    } else {
        printf("FUCK!\n");
    }
}