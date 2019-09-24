package jsonface_test

// This example shows how to use the jsonface.Unmarshal() function directly for
// advanced situations.  For normal cases, you'd use jsonface.GlobalUnmarshal() instead.

import (
    "jsonface"

    "fmt"
    "math"
    "encoding/json"
)

type (
    Transporter interface {
        Transport(distance_km float64) (time_hours float64)
    }

    Bike  struct{ NumGears int }
    Bus   struct{ LineName string }
    Tesla struct{ Charge float64 }  // Charge is a number between 0 and 1.
)

func (me Bike) Transport(distance float64) (time float64) {
    // A Bike can go at least 8 km/h, and even faster with more gears:
    return distance / ( 8 + math.Sqrt(float64(me.NumGears)) )
}
func (me Bus) Transport(distance float64) (time float64) {
    // Some bus lines are slower than others:
    var speed float64
    switch me.LineName {
    case "7":   speed=10
    case "185": speed=12
    default: panic(fmt.Errorf("Unknown Bus Line: %s",me.LineName))
    }
    return distance/speed
}
func (me Tesla) Transport(distance float64) (time float64) {
    // A Tesla goes slower as it loses charge.
    // For simplicity of this example, the car does not lose charge during transportation.
    speed := 100*me.Charge
    return distance/speed
}

func Transporter_UnmarshalJSON(bs []byte) (interface{},error) {
    var data struct { Type string }
    err := json.Unmarshal(bs, &data); if err!=nil { return nil,err }

    switch data.Type {
    case "Bike":
        var bike Bike
        err := json.Unmarshal(bs, &bike); if err!=nil { return nil,err }
        return bike,nil
    case "Bus":
        var bus Bus
        err := json.Unmarshal(bs, &bus); if err!=nil { return nil,err }
        return bus,nil
    case "Tesla":
        var tesla Tesla
        err := json.Unmarshal(bs, &tesla); if err!=nil { return nil,err }
        return tesla,nil
    default: return nil,fmt.Errorf("Unknown Transporter Type: %s",bs)
    }
}

func ExampleUnmarshal() {
    var ts []Transporter
    bs := []byte(`[{ "Type":"Bike", "NumGears":9 }, { "Type":"Bus", "LineName":"7" }]`)
    cbmap := jsonface.CBMap{ "jsonface_test.Transporter":Transporter_UnmarshalJSON }
    err := jsonface.Unmarshal(bs, &ts, cbmap); if err!=nil { panic(err) }
    fmt.Printf("%#v\n",ts)

    // Output:
    // []jsonface_test.Transporter{jsonface_test.Bike{NumGears:9}, jsonface_test.Bus{LineName:"7"}}
}

