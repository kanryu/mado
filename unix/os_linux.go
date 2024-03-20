//go:build (darwin || (linux && !baremetal && !wasi && !wasm_unknown)) && !nintendoswitch

package unix

/*
#include <time.h>
static unsigned long long get_nsecs(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (unsigned long long)ts.tv_sec * 1000000000UL + ts.tv_nsec;
}
*/
import "C"

func getTime() uint64 {
	ts := C.get_nsecs()
	return uint64(ts)
}
