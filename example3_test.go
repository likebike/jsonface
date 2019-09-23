package jsonface_test

// This is another example of direct marshalling and unmarshalling of an interface.
// In this example, the shapes have fields with the same name, therefore we need to
// add some extra type information during marshalling.

import (
    "jsonface"

    "fmt"
    "math"
    "errors"
    "encoding/json"
)

type (
    // type Shape interface { Area() float64 }  // Defined in common_test.go
    Pentagon struct { Side float64 }
    Hexagon  struct { Side float64 }
)

func (me Pentagon) Area() float64 { return (1.0/4.0)*math.Sqrt(5*(5+2*math.Sqrt(5))) * me.Side*me.Side }
func (me Hexagon) Area() float64 { return (3.0/2.0)*math.Sqrt(3) * me.Side*me.Side }

func Shape_UnmarshalJSON_2(bs []byte) (interface{},error) {
    var data struct {
        Type string
        Side float64
    }
    err:=json.Unmarshal(bs,&data); if err!=nil { return nil,err }
    if data.Side<0 { return nil,errors.New("Negative Side") }

    switch data.Type {
    case "Pentagon": return Pentagon{data.Side},nil
    case "Hexagon": return Hexagon{data.Side},nil
    default: return nil,fmt.Errorf("Unknown Shape Type: %s",bs)
    }
}

func (me Pentagon) MarshalJSON() ([]byte,error) {
    data:=struct {
        Type string
        Side float64
    }{"Pentagon",me.Side}
    return json.Marshal(data)
}
func (me Hexagon) MarshalJSON() ([]byte,error) {
    data:=struct {
        Type string
        Side float64
    }{"Hexagon",me.Side}
    return json.Marshal(data)
}

func Example_3Direct() {
    // Don't use ResetGlobalCBs in normal circumstances.  We need to use it here so our tests don't conflict:
    jsonface.ResetGlobalCBs()
    // This would normally be placed in an init() function, but I can't do that here because it conflicts with other tests:
    jsonface.AddGlobalCB("jsonface_test.Shape", Shape_UnmarshalJSON_2)

    var s1 Shape = Pentagon{5}
    var s2 Shape = Hexagon{5}
    fmt.Printf("Before: s1=%#v s2=%#v\n",s1,s2)

    s1bs,err:=json.Marshal(s1); if err!=nil { panic(err) }
    s2bs,err:=json.Marshal(s2); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: s1=%s s2=%s\n",s1bs,s2bs)

    err=jsonface.GlobalUnmarshal(s1bs,&s1); if err!=nil { panic(err) }
    err=jsonface.GlobalUnmarshal(s2bs,&s2); if err!=nil { panic(err) }
    fmt.Printf("After : s1=%#v s2=%#v\n",s1,s2)

    // Output:
    // Before: s1=jsonface_test.Pentagon{Side:5} s2=jsonface_test.Hexagon{Side:5}
    // Marshalled: s1={"Type":"Pentagon","Side":5} s2={"Type":"Hexagon","Side":5}
    // After : s1=jsonface_test.Pentagon{Side:5} s2=jsonface_test.Hexagon{Side:5}
}

