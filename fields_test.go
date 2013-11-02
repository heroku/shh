package main

import (
	"fmt"
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
	// Output: [372 (kworker/2:1) S 2 0 0 0 -1 69238880 0 0 0 0 0 17293 0 0 20 0 1 0 40 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 18446744073709551615 0 0 17 2 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleFields_parens2() {
	f := Fields("Active(anon):      42384 kB")
	fmt.Println(f)
	fmt.Println(len(f))
	// Output: [Active(anon) 42384 kB]
	// 3
}
