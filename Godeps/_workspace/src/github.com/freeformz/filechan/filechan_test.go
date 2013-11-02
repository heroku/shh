package filechan;

import (
  "fmt"
)

func ExampleFileLineChannel_basic() {
  c, _ := FileLineChannel("./filechan_fixture.txt")
  i := 0
  for line := range c {
    fmt.Print(line)
    i++
  }
  fmt.Println(i)
  //Output: Hello
  //There
  //World
  //3
}

func ExampleFileLineChannel_notFound() {
  _, err := FileLineChannel("./not_found.txt")
  if err != nil {
    fmt.Println(err)
  }
  //Output: open ./not_found.txt: no such file or directory
}

