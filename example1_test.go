package jsonface

// This is a basic example of direct marshalling and unmarshalling of an interface.
// In this particular example, the different shapes happen to have differently-named fields,
// so their types can be easily detected without adding extra type information to the marshalled data.

import (
    "fmt"
    "math"
    "errors"
    "encoding/json"
)

type (
    Shape interface { Area() float64 }
    Circle struct { Radius float64 }
    Square struct { Length float64 }
)

func (me Circle) Area() float64 { return math.Pi*me.Radius*me.Radius }
func (me Square) Area() float64 { return me.Length*me.Length }

func Shape_UnmarshalJSON(bs []byte) (interface{},error) {
    var data map[string]float64
    err:=json.Unmarshal(bs,&data); if err!=nil { return nil,err }
    if v,has:=data["Radius"]; has {
        if v<0 { return nil,errors.New("Negative Radius") }
        return Circle{v},nil
    } else if v,has:=data["Length"]; has {
        if v<0 { return nil,errors.New("Negative Length") }
        return Square{v},nil
    } else {
        return nil,fmt.Errorf("Unknown Shape Type: %s",bs)
    }
}

func init() {
    AddGlobalCB("Shape", Shape_UnmarshalJSON)
}

func Example() {
    var s1 Shape = Circle{2.5}
    var s2 Shape = Square{5.0}
    fmt.Printf("Before: s1=%#v s2=%#v\n",s1,s2)

    s1bs,err:=json.Marshal(s1); if err!=nil { panic(err) }
    s2bs,err:=json.Marshal(s2); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: s1=%s s2=%s\n",s1bs,s2bs)

    err=GlobalUnmarshal(s1bs,&s1); if err!=nil { panic(err) }
    err=GlobalUnmarshal(s2bs,&s2); if err!=nil { panic(err) }
    fmt.Printf("After: s1=%#v s2=%#v\n",s1,s2)

    // Output:
    // Before: s1=jsonface.Circle{Radius:2.5} s2=jsonface.Square{Length:5}
    // Marshalled: s1={"Radius":2.5} s2={"Length":5}
    // After: s1=jsonface.Circle{Radius:2.5} s2=jsonface.Square{Length:5}
}

