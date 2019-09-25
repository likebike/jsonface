<%! import pyhpy %>
<%block name="PAGE_CSS">
    <link rel="stylesheet" type="text/css" href="${pyhpy.url('/static/css/home.css')}">
</%block>

**Installation:** ```go get github.com/likebike/jsonface```

**Source Code:** [https://github.com/likebike/jsonface](https://github.com/likebike/jsonface)

**API docs and usage examples:** [https://godoc.org/github.com/likebike/jsonface](https://godoc.org/github.com/likebike/jsonface)


== About

jsonface enables you to isolate your data type design from your deserialization logic.

When writing Go programs, I often want to create types that contain interface members like this:

    type (
        Instrument interface {
            Play()
        }
        Bell  struct { BellPitch string }   // I am using contrived field names
        Drum  struct { DrumSize  float64 }  // to keep this example simple.
    
        BandMember struct {
            Name string
            Inst Instrument    // <---- Interface Member
        }
    )

...But if I want to serialize/deserialize a BandMember using JSON, I'm going
to have a bit of a problem because Go's json package can't unmarshal into an interface.
Therefore, I need to define some custom unmarshalling logic at the BandMember level.
This is not ideal, since the logic should really belong to Instrument, not BandMember.
It becomes especially problematic if I have other data types that also contain
Instrument members because then the unmarshalling complexity spreads there too!

This jsonface package enables me to define the unmarshalling logic at the Instrument
level, avoiding the leaky-complexity described above.

Also note, the example above just shows a very simple interface struct field,
but jsonface is very general; It can handle any data structure, no matter how
deep or complex.


== Example

```go
package main

import (
    "fmt"
    "encoding/json"

    "github.com/likebike/jsonface"
)

type (
    Instrument interface {
        Play()
    }
    Bell  struct { BellPitch string }   // I am using contrived field names
    Drum  struct { DrumSize  float64 }  // to keep this example simple.

    BandMember struct {
        Name string
        Inst Instrument    // <---- Interface Member
    }
)

func (me Bell)  Play() { fmt.Printf("Ding (%s Bell)\n", me.BellPitch) }
func (me Drum)  Play() { fmt.Printf("Boom (%f Drum)\n", me.DrumSize) }

func Instrument_UnmarshalJSON(bs []byte) (interface{},error) {
    var Keys map[string]interface{}
    err := json.Unmarshal(bs, &Keys); if err!=nil { return nil,err }

    if _,has:=Keys["BellPitch"]; has {
        var bell Bell
        err = json.Unmarshal(bs, &bell); if err!=nil { return nil,err }
        return bell,nil
    } else if _,has:=Keys["DrumSize"]; has {
        var drum Drum
        err = json.Unmarshal(bs, &drum); if err!=nil { return nil,err }
        return drum,nil
    } else {
        return nil,fmt.Errorf("Unknown Instument Type: %s",bs)
    }
}

// Register the callback with jsonface:
func init() {
    jsonface.AddGlobalCB("main.Instrument", Instrument_UnmarshalJSON)
}

func main() {
    bs := []byte(`{"Name":"Gabriella","Inst":{"BellPitch":"B♭"}}`)
    var bandmember BandMember
    err := jsonface.GlobalUnmarshal(bs,&bandmember); if err!=nil { panic(err) }
    fmt.Printf("bandmember=%#v\n",bandmember)

    // Output:
    // bandmember=main.BandMember{Name:"Gabriella", Inst:main.Bell{BellPitch:"B♭"}}
}
```

