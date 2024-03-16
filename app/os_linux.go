//go:build (darwin || (linux && !baremetal && !wasi && !wasm_unknown)) && !nintendoswitch

package app

//export clock_gettime
func libc_clock_gettime(clk_id int32, ts *timespec)

//export __clock_gettime64
func libc_clock_gettime64(clk_id int32, ts *timespec)

const intSize = 32 << (^uint(0) >> 63)

// Portable (64-bit) variant of clock_gettime.
func clock_gettime(clk_id int32, ts *timespec) {
	if intSize == 32 {
		// This is a 32-bit architecture (386, arm, etc).
		// We would like to use the 64-bit version of this function so that
		// binaries will continue to run after Y2038.
		// For more information:
		//   - https://musl.libc.org/time64.html
		//   - https://sourceware.org/glibc/wiki/Y2038ProofnessDesign
		libc_clock_gettime64(clk_id, ts)
	} else {
		// This is a 64-bit architecture (amd64, arm64, etc).
		// Use the regular variant, because it already fixes the Y2038 problem
		// by using 64-bit integer types.
		libc_clock_gettime(clk_id, ts)
	}
}

type timeUnit int64

// Note: tv_sec and tv_nsec normally vary in size by platform. However, we're
// using the time64 variant (see clock_gettime above), so the formats are the
// same between 32-bit and 64-bit architectures.
// There is one issue though: on big-endian systems, tv_nsec would be incorrect.
// But we don't support big-endian systems yet (as of 2021) so this is fine.
type timespec struct {
	tv_sec  int64 // time_t with time64 support (always 64-bit)
	tv_nsec int64 // unsigned 64-bit integer on all time64 platforms
}

func getTime(clock int32) uint64 {
	ts := timespec{}
	clock_gettime(clock, &ts)
	return uint64(ts.tv_sec)*1000*1000*1000 + uint64(ts.tv_nsec)
}
