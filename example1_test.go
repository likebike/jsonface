package jsonface_test

// This example shows how you might normally design your data types and serialization
// logic without jsonface, and then it shows how to do the same thing with jsonface.

import (
    "jsonface"

    "fmt"
    "encoding/json"
)

type (
    Instrument interface {
        Play()
    }
    Bell  struct { Size float64 }
    Drum  struct { Size float64 }

    BandMember_NoJsonface struct {
        Name       string
        Instrument Instrument
    }

    BandMember_UsingJsonface struct {
        Name       string
        Instrument Instrument
    }
)

func (me Bell)  Play() { fmt.Printf("Ding (%f Bell)\n", me.Size) }
func (me Drum)  Play() { fmt.Printf("Boom (%f Drum)\n", me.Size) }

// Define some custom Marshalling behavior so we can add type info:
func (me Bell) MarshalJSON() ([]byte,error) {
    data := struct {
        Type string
        Size float64
    }{"Bell", me.Size}
    return json.Marshal(data)
}
func (me Drum) MarshalJSON() ([]byte,error) {
    data := struct {
        Type string
        Size float64
    }{"Drum", me.Size}
    return json.Marshal(data)
}

// This is the normal solution for this situation; the Instrument unmarshalling complexity
// leaks out to the BandMember level:
func (me *BandMember_NoJsonface) UnmarshalJSON(bs []byte) error {
    var data struct {
        Name       string
        Instrument json.RawMessage
    }
    err := json.Unmarshal(bs, &data); if err!=nil { return err }

    var InstrumentType struct { Type string }
    err = json.Unmarshal(data.Instrument, &InstrumentType); if err!=nil { return err }

    var InstrumentObj Instrument
    switch InstrumentType.Type {
    case "Bell":
        var bell Bell
        err = json.Unmarshal(data.Instrument, &bell); if err!=nil { return err }
        InstrumentObj = bell
    case "Drum":
        var drum Drum
        err = json.Unmarshal(data.Instrument, &drum); if err!=nil { return err }
        InstrumentObj = drum
    default: return fmt.Errorf("Unknown Instument Type: %s",data.Instrument)
    }

    me.Name, me.Instrument = data.Name, InstrumentObj
    return nil
}

// This is the jsonface version of the above function.  It contains the complexity within the Instrument type.
func Instrument_UnmarshalJSON(bs []byte) (interface{},error) {

    var InstrumentType struct { Type string }
    err := json.Unmarshal(bs, &InstrumentType); if err!=nil { return nil,err }

    switch InstrumentType.Type {
    case "Bell":
        var bell Bell
        err = json.Unmarshal(bs, &bell); if err!=nil { return nil,err }
        return bell,nil
    case "Drum":
        var drum Drum
        err = json.Unmarshal(bs, &drum); if err!=nil { return nil,err }
        return drum,nil
    default: return nil,fmt.Errorf("Unknown Instument Type: %s",bs)
    }
}

// Register the jsonface callback:
func init() {
    jsonface.AddGlobalCB("jsonface_test.Instrument", Instrument_UnmarshalJSON)
}

func Example_1BeforeAfter() {
    // An example without jsonface:
    member1 := BandMember_NoJsonface{ "Christopher", Drum{25} }
    fmt.Printf("Before: member1=%#v\n",member1)
    m1bs,err := json.Marshal(member1); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: member1=%s\n",m1bs)
    err = json.Unmarshal(m1bs,&member1); if err!=nil { panic(err) }
    fmt.Printf("After : member1=%#v\n",member1)

    // An example with jsonface:
    member2 := BandMember_UsingJsonface{ "Gabriella", Bell{10} }
    fmt.Printf("Before: member2=%#v\n",member2)
    m2bs,err := json.Marshal(member2); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: member2=%s\n",m2bs)
    err = jsonface.GlobalUnmarshal(m2bs,&member2); if err!=nil { panic(err) }
    fmt.Printf("After : member2=%#v\n",member2)

    // Output:
    // Before: member1=jsonface_test.BandMember_NoJsonface{Name:"Christopher", Instrument:jsonface_test.Drum{Size:25}}
    // Marshalled: member1={"Name":"Christopher","Instrument":{"Type":"Drum","Size":25}}
    // After : member1=jsonface_test.BandMember_NoJsonface{Name:"Christopher", Instrument:jsonface_test.Drum{Size:25}}
    // Before: member2=jsonface_test.BandMember_UsingJsonface{Name:"Gabriella", Instrument:jsonface_test.Bell{Size:10}}
    // Marshalled: member2={"Name":"Gabriella","Instrument":{"Type":"Bell","Size":10}}
    // After : member2=jsonface_test.BandMember_UsingJsonface{Name:"Gabriella", Instrument:jsonface_test.Bell{Size:10}}
}

