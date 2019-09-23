package jsonface_test

import (
    "jsonface"

    "fmt"
    "os"
)

type A int64

func ExampleGetTypeName() {
    var i int64
    fmt.Println("i:",jsonface.GetTypeName(i))

    var a A
    fmt.Println("a:",jsonface.GetTypeName(a))

    fmt.Println("&a:",jsonface.GetTypeName(&a))

    fmt.Println("os.Stdout:",jsonface.GetTypeName(os.Stdout))

    // Output:
    // i: int64
    // a: jsonface_test.A
    // &a: *jsonface_test.A
    // os.Stdout: *os.File
}

