package utils

import (
	"fmt"
	"os"
)

// GetEnvWithDefault

func ExampleGetEnvWithDefault_notSet() {
	os.Clearenv()
	fmt.Println(GetEnvWithDefault("SHH_TEST_ENV", "bar"))
	// Output: bar
}

func ExampleGetEnvWithDefault_empty() {
	os.Setenv("SHH_TEST_ENV", "")
	fmt.Println(GetEnvWithDefault("SHH_TEST_ENV", "bar"))
	// Output:
}

func ExampleGetEnvWithDefault_notDefault() {
	os.Setenv("SHH_TEST_ENV", "foo")
	fmt.Println(GetEnvWithDefault("SHH_TEST_ENV", "bar"))
	// Output: foo
}

func ExampleGetEnvWithDefault_default() {
	os.Setenv("SHH_TEST_ENV", "bar")
	fmt.Println(GetEnvWithDefault("SHH_TEST_ENV", "bar"))
	// Output: bar
}

// GetEnvWithDefaultInt

func ExampleGetEnvWithDefaultInt_notSet() {
	os.Clearenv()
	fmt.Println(GetEnvWithDefaultInt("SHH_TEST_ENV", 42))
	// Output: 42
}

func ExampleGetEnvWithDefaultInt_empty() {
	os.Setenv("SHH_TEST_ENV", "")
	fmt.Println(GetEnvWithDefaultInt("SHH_TEST_ENV", 42))
	// Output: 42
}

func ExampleGetEnvWithDefaultInt_notDefault() {
	os.Setenv("SHH_TEST_ENV", "7")
	fmt.Println(GetEnvWithDefaultInt("SHH_TEST_ENV", 42))
	// Output: 7
}

func ExampleGetEnvWithDefaultInt_default() {
	os.Setenv("SHH_TEST_ENV", "42")
	fmt.Println(GetEnvWithDefaultInt("SHH_TEST_ENV", 42))
	// Output: 42
}

// GetEnvWithDefaultDuration

func ExampleGetEnvWithDefaultDuration_notSet() {
	os.Clearenv()
	fmt.Println(GetEnvWithDefaultDuration("SHH_TEST_ENV", "42s"))
	// Output: 42s
}

func ExampleGetEnvWithDefaultDuration_empty() {
	os.Setenv("SHH_TEST_ENV", "")
	fmt.Println(GetEnvWithDefaultDuration("SHH_TEST_ENV", "42s"))
	// Output: 42s
}

func ExampleGetEnvWithDefaultDuration_notDefault() {
	os.Setenv("SHH_TEST_ENV", "7s")
	fmt.Println(GetEnvWithDefaultDuration("SHH_TEST_ENV", "42s"))
	// Output: 7s
}

func ExampleGetEnvWithDefaultDuration_default() {
	os.Setenv("SHH_TEST_ENV", "42s")
	fmt.Println(GetEnvWithDefaultDuration("SHH_TEST_ENV", "42s"))
	// Output: 42s
}

// GetEnvWithDefaultStrings

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
