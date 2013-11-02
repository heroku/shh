package main

import (
	"fmt"
	"time"
)

// Slog
func ExampleSlog_single() {
	fmt.Println(Slog{"foo": "bar"})
	//Output: foo=bar
}

func ExampleSlog_single_num() {
	fmt.Println(Slog{"foo": 1})
	//Output: foo=1
}

func ExampleSlog_multi() {
	fmt.Println(Slog{"foo": "bar", "bar": "bazzle"})

	//Output: bar=bazzle foo=bar
}

func ExampleSlog_mixed() {
	fmt.Println(Slog{"foo": "bar", "bar": "bazzle", "baz": 1, "bazzle": true})

	//Output: bar=bazzle baz=1 bazzle=true foo=bar
}

func ExampleSlog_withError() {
	fmt.Println(Slog{"error": fmt.Errorf("fi fie fo fum")})
	//Output: error="fi fie fo fum"
}

func ExampleSlog_add() {
	l := Slog{"start": true}
	fmt.Println(l)

	l["error"] = "BOOM"
	fmt.Println(l)

	//Output: start=true
	// error=BOOM start=true
}

func ExampleSlog_time() {
	fmt.Println(Slog{"now": time.Unix(0, 0).UTC()})
	//Output: now="1970-01-01T00:00:00Z"
}

func ExampleSlog_replace() {
	l := Slog{"start": "here"}
	fmt.Println(l)

	l["start"] = "there"
	fmt.Println(l)

	//Output: start=here
	// start=there
}
