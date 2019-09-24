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
    Bell  struct { BellPitch string }  // I am using contrived field names
    Drum  struct { DrumSize  float64 } // to keep this example simple.

    BandMember struct {
        Name string
        Inst Instrument
    }

    BandMember_UsingJsonface BandMember
)

func (me Bell)  Play() { fmt.Printf("Ding (%s Bell)\n", me.BellPitch) }
func (me Drum)  Play() { fmt.Printf("Boom (%f Drum)\n", me.DrumSize) }

// This is the normal solution for this situation; the Instrument unmarshalling
// complexity leaks out to the BandMember level:
func (me *BandMember) UnmarshalJSON(bs []byte) error {
    var data struct {
        Name string
        Inst json.RawMessage
    }
    err := json.Unmarshal(bs, &data); if err!=nil { return err }

    var Keys map[string]interface{}
    err = json.Unmarshal(data.Inst, &Keys); if err!=nil { return err }

    var InstrumentObj Instrument
    if _,has:=Keys["BellPitch"]; has {
        var bell Bell
        err = json.Unmarshal(data.Inst, &bell); if err!=nil { return err }
        InstrumentObj = bell
    } else if _,has:=Keys["DrumSize"]; has {
        var drum Drum
        err = json.Unmarshal(data.Inst, &drum); if err!=nil { return err }
        InstrumentObj = drum
    } else {
        return fmt.Errorf("Unknown Instument Type: %s",data.Inst)
    }

    me.Name, me.Inst = data.Name, InstrumentObj
    return nil
}

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

// Register the jsonface callback:
func init() {
    jsonface.AddGlobalCB("jsonface_test.Instrument", Instrument_UnmarshalJSON)
}

func Example_1BeforeAfter() {
    // An example without jsonface:
    member1 := BandMember{ "Christopher", Drum{25} }
    fmt.Printf("Before: member1=%#v\n",member1)
    m1bs,err := json.Marshal(member1); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: member1=%s\n",m1bs)
    err = json.Unmarshal(m1bs,&member1); if err!=nil { panic(err) }
    fmt.Printf("After : member1=%#v\n",member1)

    // An example with jsonface:
    member2 := BandMember_UsingJsonface{ "Gabriella", Bell{"B♭"} }
    fmt.Printf("Before: member2=%#v\n",member2)
    m2bs,err := json.Marshal(member2); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: member2=%s\n",m2bs)
    err = jsonface.GlobalUnmarshal(m2bs,&member2); if err!=nil { panic(err) }
    fmt.Printf("After : member2=%#v\n",member2)

    // Output:
    // Before: member1=jsonface_test.BandMember{Name:"Christopher", Inst:jsonface_test.Drum{DrumSize:25}}
    // Marshalled: member1={"Name":"Christopher","Inst":{"DrumSize":25}}
    // After : member1=jsonface_test.BandMember{Name:"Christopher", Inst:jsonface_test.Drum{DrumSize:25}}
    // Before: member2=jsonface_test.BandMember_UsingJsonface{Name:"Gabriella", Inst:jsonface_test.Bell{BellPitch:"B♭"}}
    // Marshalled: member2={"Name":"Gabriella","Inst":{"BellPitch":"B♭"}}
    // After : member2=jsonface_test.BandMember_UsingJsonface{Name:"Gabriella", Inst:jsonface_test.Bell{BellPitch:"B♭"}}
}

