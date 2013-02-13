package utils

import (
	"fmt"
)

func ExampleAtouint64_small() {
	fmt.Println(Atouint64("0"))
	// Output: 0
}

func ExampleAtouint64_big() {
	fmt.Println(Atouint64("10226292680"))
	// Output: 10226292680
}

func ExampleUi64toa() {
	fmt.Println(Ui64toa(10226292680))
	// Output: 10226292680
}

func ExampleAtofloat64_small() {
	fmt.Println(Atofloat64("0.0"))
	// Output: 0
}

func ExampleAtofloat64_big() {
	fmt.Println(Atofloat64("10226292680.3"))
	// Output: 1.02262926803e+10
}

func ExamplePercentFormat() {
	fmt.Println(PercentFormat(100.0))
	fmt.Println(PercentFormat(42.314))
	// Output: 100.00
	// 42.31
}
