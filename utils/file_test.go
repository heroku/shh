package utils

import (
	"fmt"
)

func ExampleExists() {
	fmt.Println(Exists("./foozle_not_found"))
	fmt.Println(Exists("./file_test.go"))
	// Output: false
	// true
}
