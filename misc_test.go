package main

import (
	"fmt"
)

// FileLineChannel
func ExampleFileLineChannel_noError() {
	c := FileLineChannel("./misc_test.go")
	var i int
	for d := range c {
		d = d
		i++
	}
	fmt.Println(i > 0)
	//Output: true
}

//FixUpName
func ExampleFixUpName() {
	fmt.Println(FixUpName("foo(bar)"))
	fmt.Println(FixUpName("Foo(Bar)"))
	fmt.Println(FixUpName("foo_bar"))
	fmt.Println(FixUpName("Foo_Bar)"))
	fmt.Println(FixUpName("Foo"))
	fmt.Println(FixUpName("(Foo)"))
	//Output: [foo bar]
	//[foo bar]
	//[foo bar]
	//[foo bar]
	//[foo]
	//[foo]
}
