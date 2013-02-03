package utils

import "fmt"

func ExampleFields_basic() {
	fmt.Println(Fields("  eth0: 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0"))
	// Output: [eth0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
}

func ExampleFields_squashed() {
	fmt.Println(Fields("  eth0:10226292680 39079204    0    0    0     0          0         0 10250230999 51012120    0    0    0     0       0          0\n"))
	// Output: [eth0 10226292680 39079204 0 0 0 0 0 0 10250230999 51012120 0 0 0 0 0 0]
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
