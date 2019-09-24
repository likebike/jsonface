package jsonface_test

// In this example, we show that jsonface supports unmarshalling of interfaces
// that are stored in arbitrary composite types, such as structs, slices, and maps.
// jsonface is completely recursive and general, so every data type is supported,
// no matter how complex.

import (
    "jsonface"

    "fmt"
    "encoding/json"
)

type (
    Food interface{}

    Water      struct{}
    Ice        Water               // Water can transform into Ice.
    Grass      struct{W Water}     // A Grass consumes 1 Water.
    Corn       struct{Ws []Water}  // A Corn consumes multiple Waters.
    Cornflakes struct{C Corn}      // Cornflakes is made of Corn.
    Cow        struct {            // Cow is a Food that also eats Foods.
                   Name string
                   Ate  []Food
               }
    Milk       struct{}
    Cream      Milk     // Milk can transform into Cream.
    IceCream   struct { // IceCream consumes 1 Ice and 1 Cream.
                   I Ice
                   C Cream
               }
    MealName   string   // Breakfast, Lunch, Dinner, etc.
    Girl       struct {
                   Name  string
                   Meals map[MealName][]Food
               }
)

func Food_UnmarshalJSON(bs []byte) (interface{},error) {
    var data struct { Type string }
    err := json.Unmarshal(bs,&data); if err!=nil { return nil,err }

    switch data.Type {
    case "Water": return Water{},nil
    case "Ice": return Ice{},nil
    case "Grass":
        var x Grass
        err = jsonface.GlobalUnmarshal(bs,&x); if err!=nil { return nil,err }
        return x,nil
    case "Corn":
        var x Corn
        err = jsonface.GlobalUnmarshal(bs,&x); if err!=nil { return nil,err }
        return x,nil
    case "Cornflakes":
        var x Cornflakes
        err = jsonface.GlobalUnmarshal(bs,&x); if err!=nil { return nil,err }
        return x,nil
    case "Cow":
        // The Cow type contains nested Food interfaces which must also be unmarshalled.
        type X Cow  // Use indirection to avoid infinite recursion.
        var x X
        err = jsonface.GlobalUnmarshal(bs,&x); if err!=nil { return nil,err }
        return Cow(x),nil
    case "Milk": return Milk{},nil
    case "Cream": return Cream{},nil
    case "IceCream":
        var x IceCream
        err = jsonface.GlobalUnmarshal(bs,&x); if err!=nil { return nil,err }
        return x,nil
    default: return nil,fmt.Errorf("Unknown Food Type: %s",bs)
    }
}

func (me Water) MarshalJSON() ([]byte,error) {
    data := struct{Type string}{"Water"}
    return json.Marshal(data)
}
func (me Ice) MarshalJSON() ([]byte,error) {
    data := struct{Type string}{"Ice"}
    return json.Marshal(data)
}
func (me Grass) MarshalJSON() ([]byte,error) {
    data := struct{
        Type string
        W    Water
    }{"Grass",me.W}
    return json.Marshal(data)
}
func (me Corn) MarshalJSON() ([]byte,error) {
    data := struct{
        Type string
        Ws   []Water
    }{"Corn",me.Ws}
    return json.Marshal(data)
}
func (me Cornflakes) MarshalJSON() ([]byte,error) {
    data := struct{
        Type string
        C    Corn
    }{"Cornflakes",me.C}
    return json.Marshal(data)
}
func (me Cow) MarshalJSON() ([]byte,error) {
    data := struct{
        Type string
        Name string
        Ate  []Food
    }{"Cow",me.Name,me.Ate}
    return json.Marshal(data)
}
func (me Milk) MarshalJSON() ([]byte,error) {
    data := struct{Type string}{"Milk"}
    return json.Marshal(data)
}
func (me Cream) MarshalJSON() ([]byte,error) {
    data := struct{Type string}{"Cream"}
    return json.Marshal(data)
}
func (me IceCream) MarshalJSON() ([]byte,error) {
    data := struct{
        Type string
        I    Ice
        C    Cream
    }{"IceCream",me.I,me.C}
    return json.Marshal(data)
}

func Example_4Composites() {
    // Don't use ResetGlobalCBs in normal circumstances.  We need to use it here
    // so our tests don't conflict:
    jsonface.ResetGlobalCBs()
    // These would normally be placed in an init() function, but I can't do that
    // here because it conflicts with other tests:
    jsonface.AddGlobalCB("jsonface_test.Food", Food_UnmarshalJSON)

    // It rains.  10 Waters are produced:
    waters := make([]Water,10)
    water := func() (w Water) {
        w,waters = waters[len(waters)-1],waters[:len(waters)-1]
        return
    }

    // Some Grass and Corn grows:
    grass := Grass{water()}
    corn1 := Corn{[]Water{ water(),water() }}
    corn2 := Corn{[]Water{ water(),water(),water() }}

    // One Corn is turned into Cornflakes:
    cornflakes := Cornflakes{corn1}

    // The cow eats Grass, Corn, and Water:
    cow := Cow{"Bessie", []Food{ grass,corn2,water() }}

    // The cow produces one Milk for each Food it ate:
    milks := make([]Milk,len(cow.Ate))
    milk := func() (m Milk) {
        m,milks = milks[len(milks)-1],milks[:len(milks)-1]
        return
    }

    // One Milk is turned into Cream:
    cream := Cream(milk())

    // One Water is turned into Ice:
    ice := Ice(water())

    // Make IceCream:
    icecream := IceCream{ice,cream}

    // Gabriella eats cereal for breakfast, icecream for lunch, and steak for dinner:
    gabriella := Girl{"Gabriella", map[MealName][]Food{
        "Breakfast":{ cornflakes,milk() },
        "Lunch":{ icecream,water() },
        "Dinner":{ cow,milk() },
    }}
    fmt.Printf("Before: gabriella=%#v\n",gabriella)

    bs,err := json.Marshal(gabriella); if err!=nil { panic(err) }
    fmt.Printf("Marshalled: gabriella=%s\n",bs)

    err = jsonface.GlobalUnmarshal(bs,&gabriella); if err!=nil { panic(err) }
    fmt.Printf("After : gabriella=%#v\n",gabriella)

    // Output:
    // Before: gabriella=jsonface_test.Girl{Name:"Gabriella", Meals:map[jsonface_test.MealName][]jsonface_test.Food{"Breakfast":[]jsonface_test.Food{jsonface_test.Cornflakes{C:jsonface_test.Corn{Ws:[]jsonface_test.Water{jsonface_test.Water{}, jsonface_test.Water{}}}}, jsonface_test.Milk{}}, "Dinner":[]jsonface_test.Food{jsonface_test.Cow{Name:"Bessie", Ate:[]jsonface_test.Food{jsonface_test.Grass{W:jsonface_test.Water{}}, jsonface_test.Corn{Ws:[]jsonface_test.Water{jsonface_test.Water{}, jsonface_test.Water{}, jsonface_test.Water{}}}, jsonface_test.Water{}}}, jsonface_test.Milk{}}, "Lunch":[]jsonface_test.Food{jsonface_test.IceCream{I:jsonface_test.Ice{}, C:jsonface_test.Cream{}}, jsonface_test.Water{}}}}
    // Marshalled: gabriella={"Name":"Gabriella","Meals":{"Breakfast":[{"Type":"Cornflakes","C":{"Type":"Corn","Ws":[{"Type":"Water"},{"Type":"Water"}]}},{"Type":"Milk"}],"Dinner":[{"Type":"Cow","Name":"Bessie","Ate":[{"Type":"Grass","W":{"Type":"Water"}},{"Type":"Corn","Ws":[{"Type":"Water"},{"Type":"Water"},{"Type":"Water"}]},{"Type":"Water"}]},{"Type":"Milk"}],"Lunch":[{"Type":"IceCream","I":{"Type":"Ice"},"C":{"Type":"Cream"}},{"Type":"Water"}]}}
    // After : gabriella=jsonface_test.Girl{Name:"Gabriella", Meals:map[jsonface_test.MealName][]jsonface_test.Food{"Breakfast":[]jsonface_test.Food{jsonface_test.Cornflakes{C:jsonface_test.Corn{Ws:[]jsonface_test.Water{jsonface_test.Water{}, jsonface_test.Water{}}}}, jsonface_test.Milk{}}, "Dinner":[]jsonface_test.Food{jsonface_test.Cow{Name:"Bessie", Ate:[]jsonface_test.Food{jsonface_test.Grass{W:jsonface_test.Water{}}, jsonface_test.Corn{Ws:[]jsonface_test.Water{jsonface_test.Water{}, jsonface_test.Water{}, jsonface_test.Water{}}}, jsonface_test.Water{}}}, jsonface_test.Milk{}}, "Lunch":[]jsonface_test.Food{jsonface_test.IceCream{I:jsonface_test.Ice{}, C:jsonface_test.Cream{}}, jsonface_test.Water{}}}}
}

