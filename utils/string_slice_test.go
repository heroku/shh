package utils

import (
	"fmt"
)

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

func ExampleLinearSliceContainsString() {
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
