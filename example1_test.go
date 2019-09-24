package jsonface_test

// This example shows how you might normally design your data types and
// serialization logic without jsonface, and then it shows how to do the
// same thing with jsonface.

import (
    "jsonface"

    "fmt"
    "encoding/json"
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

// // This is the normal solution for this situation; the Instrument unmarshalling
// // complexity leaks out to the BandMember level:
// func (me *BandMember) UnmarshalJSON(bs []byte) error {
//     var data struct {
//         Name string
//         Inst json.RawMessage
//     }
//     err := json.Unmarshal(bs, &data); if err!=nil { return err }
// 
//     var Keys map[string]interface{}
//     err = json.Unmarshal(data.Inst, &Keys); if err!=nil { return err }
// 
//     var InstrumentObj Instrument
//     if _,has:=Keys["BellPitch"]; has {
//         var bell Bell
//         err = json.Unmarshal(data.Inst, &bell); if err!=nil { return err }
//         InstrumentObj = bell
//     } else if _,has:=Keys["DrumSize"]; has {
//         var drum Drum
//         err = json.Unmarshal(data.Inst, &drum); if err!=nil { return err }
//         InstrumentObj = drum
//     } else {
//         return fmt.Errorf("Unknown Instument Type: %s",data.Inst)
//     }
// 
//     me.Name, me.Inst = data.Name, InstrumentObj
//     return nil
// }

// This is the jsonface version of the above function.  It contains the complexity
// within the Instrument type.
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
    jsonface.AddGlobalCB("jsonface_test.Instrument", Instrument_UnmarshalJSON)
}

func Example_1BeforeAfter() {
    var bandmember BandMember
    bs := []byte(`{"Name":"Gabriella","Inst":{"BellPitch":"B♭"}}`)
    err := jsonface.GlobalUnmarshal(bs,&bandmember); if err!=nil { panic(err) }
    fmt.Printf("bandmember=%#v\n",bandmember)

    // Output:
    // bandmember=jsonface_test.BandMember{Name:"Gabriella", Inst:jsonface_test.Bell{BellPitch:"B♭"}}
}

