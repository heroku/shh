package utils

import (
	"fmt"
	"os"
)

func ExampleFields_basic() {
	fmt.Println(Fields("  eth0: 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"))
	// Output: [eth0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleFields_squashed() {
	fmt.Println(Fields("  eth0:10226292680 39079204    0    0    0     0          0         0 10250230999 51012120    0    0    0     0       0          0\n"))
	// Output: [eth0 10226292680 39079204 0 0 0 0 0 0 10250230999 51012120 0 0 0 0 0 0]
}

func ExampleFields_parens() {
	fmt.Println(Fields("372 (kworker/2:1) S 2 0 0 0 -1 69238880 0 0 0 0 0 17293 0 0 20 0 1 0 40 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 2 0 0 0 0 0 0 0 0 0 0 0 0 0\n"))
	// Output: [372 kworker/2:1 S 2 0 0 0 -1 69238880 0 0 0 0 0 17293 0 0 20 0 1 0 40 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 2 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleFields_parens2() {
	f := Fields("Active(anon):      42384 kB")
	fmt.Println(f)
	fmt.Println(len(f))
	// Output: [Active anon 42384 kB]
	// 4
}

func ExampleAtouint64_small() {
	fmt.Println(Atouint64("0"))
	// Output: 0
}

func ExampleAtouint64_big() {
	fmt.Println(Atouint64("10226292680"))
	// Output: 10226292680
}

func ExampleSliceContainsString() {
	fmt.Println(SliceContainsString([]string{"a", "b", "c"}, "a"))
	fmt.Println(SliceContainsString([]string{"a", "b", "c"}, "b"))
	fmt.Println(SliceContainsString([]string{"a", "b", "c"}, "c"))
	fmt.Println(SliceContainsString([]string{"a", "b", "c"}, "z"))
	fmt.Println(SliceContainsString([]string{"a", "b", "c"}, "aa"))
	// Output: true
	// true
	// true
	// false
	// false
}

func ExampleGetEnvWithDefaultStrings_empty() {
	os.Setenv("SHH_TEST_ENV", "")
	fmt.Println(len(GetEnvWithDefaultStrings("SHH_TEST_ENV", "")))
	// Output: 0
}

func ExampleGetEnvWithDefaultStrings_single() {
	os.Setenv("SHH_TEST_ENV", "foo")
	v := GetEnvWithDefaultStrings("SHH_TEST_ENV", "")
	fmt.Println(len(v))
	fmt.Println(v[0])
	// Output: 1
	// foo
}

func ExampleGetEnvWithDefaultStrings_multiple() {
	os.Setenv("SHH_TEST_ENV", "foo,bar")
	v := GetEnvWithDefaultStrings("SHH_TEST_ENV", "")
	fmt.Println(len(v))
	fmt.Println(v[0])
	fmt.Println(v[1])
	// Output: 2
	// bar
	// foo
}
