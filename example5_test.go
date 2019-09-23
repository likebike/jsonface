package jsonface_test

import (
    "fmt"
    "os"
)

type A int64

func ExampleGetTypeName() {
    var i int64
    fmt.Println("i:",GetTypeName(i))

    var a A
    fmt.Println("a:",GetTypeName(a))

    fmt.Println("&a:",GetTypeName(&a))

    fmt.Println("os.Stdout:",GetTypeName(os.Stdout))

    // Output:
    // i: int64
    // a: jsonface_test.A
    // &a: *jsonface_test.A
    // os.Stdout: *os.File
}

