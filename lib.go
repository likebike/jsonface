// Copyright 2019 Christopher Sebastian.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// Package jsonface enables JSON Unmarshalling into Go Interfaces.
// This enables you to isolate your data type design from your deserialization logic.
//
// When writing Go programs, I often want to create types that contain interface members like this:
//
//     type (
//         Instrument interface {
//             Play()
//         }
//         Bell  struct { BellPitch string }   // I am using contrived field names
//         Drum  struct { DrumSize  float64 }  // to keep this example simple.
//     
//         BandMember struct {
//             Name string
//             Inst Instrument    // <---- Interface Member
//         }
//     )
//
// ...But if I want to serialize/deserialize a BandMember using JSON, I'm going
// to have a bit of a problem because Go's json package can't unmarshal into an interface.
// Therefore, I need to define some custom unmarshalling logic at the BandMember level.
// This is not ideal, since the logic should really belong to Instrument, not BandMember.
// It becomes especially problematic if I have other data types that also contain
// Instrument members because then the unmarshalling complexity spreads there too!
//
// This jsonface package enables me to define the unmarshalling logic at the Instrument
// level, avoiding the leaky-complexity described above.
//
// Also note, the example above just shows a very simple interface struct field,
// but jsonface is very general; It can handle any data structure, no matter how
// deep or complex.
//
// See the included examples for more usage information.
package jsonface

import (
    "fmt"
    "os"
    "errors"
    "reflect"
    "encoding"
    "encoding/json"
    "sync"
)

// 'CB' means 'Callback'.  It is used for unmarshalling, with the same interface
// as an UnmarshalJSON method.
type CB func([]byte) (interface{},error)

// TypeName is the name of a type (usually prefixed by the package name).
// If you don't know the correct TypeName to use, try the GetTypeName() function.
type TypeName string

// CBMap is a TypeName-->CB mapping.  It is used to tell the jsonface system which
// callbacks to use for which types.
type CBMap map[TypeName]CB

// GetTypeName can help you understand the correct TypeNames to use during development.
// After you understand how the TypeNames are made, you will usually just hard-code the
// names into your code, rather than using this function.
//
// Coincidentally, this function produces the same result as fmt.Sprintf("%T",x) .
func GetTypeName(x interface{}) TypeName {
    return TypeName(reflect.TypeOf(x).String())   // String() is more precise than Name().
}

// We want to be able to propagate CB-Generated errors directly.
// This cbErr type allows us to detect CB-Generated errors vs our own-generated errors:
type cbErr struct { e error }
func (me cbErr) Error() string { return "This is a jsonface.cbErr; it should be unwrapped." }
func fmtErr(msg string, e error) error {
    if e==nil { return nil }
    switch E:=e.(type) {
    case cbErr: return E
    default: return fmt.Errorf(msg,e)
    }
}
func unwrapCBErr(e error) error {
    if e==nil { return nil }
    switch E:=e.(type) {
    case cbErr: return E.e
    default: return e
    }
}

// StuntDouble is a type used internally within jsonface.  Users of jsonface
// should ignore this type.  It is an exported symbol (capitalized) for
// technical reasons -- the Go json unmarshaller requires destination types to
// be exported; an unexported symbol (lowercase) would not work.
// I apologize for the API noise.
type StuntDouble string
func (me StuntDouble) MarshalJSON() ([]byte,error) {
    if len(me)==0 { return []byte("null"),nil }
    return []byte(me),nil
}
func (me *StuntDouble) UnmarshalJSON(bs []byte) error {
    if me==nil { return errors.New("jsonface.StuntDouble: UnmarshalJSON on nil pointer") }
    *me=StuntDouble(bs)
    return nil
}

var _STUNT_TYPE=reflect.TypeOf(StuntDouble(""))
var _JSON_UNMARSHALER_TYPE=reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
var _TEXT_UNMARSHALER_TYPE=reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

var globalCBs=struct {
    sync.RWMutex
    m CBMap
}{sync.RWMutex{},CBMap{}}

// AddGlobalCB adds an entry to the global callback registry.
// Then, when GlobalUnmarshal() is called, this global registry will be used to
// perform the unmarshalling.  You will normally call AddGlobalCB() during
// program initialization (from an init() function) to register your
// unmarshallable interfaces.
func AddGlobalCB(name TypeName, cb CB) {
    globalCBs.Lock(); defer globalCBs.Unlock()
    if _,has:=globalCBs.m[name]; has { panic(errors.New("CB already defined")) }
    globalCBs.m[name]=cb
}

// ResetGlobalCBs removes all definitions from the global callback registry.
// You probably shouldn't use this -- I just need to use it from my unit tests
// because Go runs all tests consecutively without resetting the namespace, and
// so my tests conflict with eachother.  I need to use this to reset the
// registry between tests.
//
// If you think you need this, instead consider using Unmarshal() and passing
// in your own CBMap.
func ResetGlobalCBs() {
    fmt.Fprintln(os.Stderr, "Warning: You are calling ResetGlobalCBs.  This should probably only be used from the jsonface unit tests!")
    globalCBs.Lock(); defer globalCBs.Unlock()
    for k:=range globalCBs.m { delete(globalCBs.m,k) }
}

// GlobalUnmarshal uses the global callback registry (created by the
// AddGlobalCB() funcion) to unmarshal data.
func GlobalUnmarshal(bs []byte, destPtr interface{}) error {
    globalCBs.RLock(); defer globalCBs.RUnlock()
    return Unmarshal(bs,destPtr,globalCBs.m)
}

// Unmarshal uses the provided CBMap to perform unmarshalling.  It does not use
// the global callback registry.  Most users will want to use GlobalUnmarshal()
// instead, but this function is provided for extra flexibility in advanced
// situations.
//
// Some "advanced situations" where you might want to use Unmarshal() are:
//
//     * You want to unmarshal many objects in parallel.  (GlobalUnmarshal
//       uses a lock, and therefore only processes items in series.)
//
//     * You only want the callback registration to be temporary.
//
//     * You are creating and *destroying* types dynamically.
//
//     * You need to avoid name collisions.  (Not usually a problem.)
func Unmarshal(bs []byte, destPtr interface{}, cbs CBMap) error {
    destPtrV:=reflect.ValueOf(destPtr)
    if !destPtrV.IsValid() { return errors.New("invalid destPtr") }
    if destPtrV.Kind()!=reflect.Ptr { return errors.New("destPtr is not a pointer") }
    if destPtrV.IsNil() { return errors.New("nil destPtr") }
    destType:=destPtrV.Elem().Type(); if destType==nil { return errors.New("nil destType") }
    sdType,hasStunt,e:=stuntdoubleType(destType,cbs); if e!=nil { return fmt.Errorf("stuntdoubleType error: %v",e) }
    if !hasStunt { return json.Unmarshal(bs,destPtr) }  // If no stunt was used, just fallback to standard behavior.
    sdPtrV:=reflect.New(sdType)
    if !sdPtrV.CanInterface() { return errors.New("cannot sdPtrV.Interface()") }
    e=json.Unmarshal(bs,sdPtrV.Interface()); if e!=nil { return fmt.Errorf("json.Unmarshal error: %v",e) }
    e=stuntdoubleToReal(sdPtrV,destPtrV,cbs); if e!=nil { return unwrapCBErr(fmtErr("stuntdoubleToReal error: %v",e)) }
    return nil
}

// stuntdoubleType transforms the given 'realType' to a StuntDouble type.
// Primitive types (like int) and types that do not have an entry in the CBMap
// do not need transformation, and are returned directly.
func stuntdoubleType(realType reflect.Type, cbs CBMap) (reflect.Type,bool,error) {
    if realType==nil { return nil,false,errors.New("nil realType!  If you are trying to get the type of an interface, you must use some indirection because Go discards the types of interface values at compile time.  See https://golang.org/pkg/reflect/#TypeOf .  Example: var x MyInterface; stuntdoubleType(reflect.ValueOf(&x).Elem().Type(), cbs)") }

    // Check realType and its pointer type for Unmarshaler:
    realPtrType:=reflect.PtrTo(realType)
    if realType.Implements(_JSON_UNMARSHALER_TYPE) || realPtrType.Implements(_JSON_UNMARSHALER_TYPE) ||
       realType.Implements(_TEXT_UNMARSHALER_TYPE) || realPtrType.Implements(_TEXT_UNMARSHALER_TYPE) { return realType,false,nil }  // Don't descend into this type to avoid losing the custom unmarshaling behavior which probably sets unexported fields that we wouldn't be able to access.

    switch realType.Kind() {
    case reflect.Invalid:
        return nil,false,errors.New("invalid kind")
    case reflect.Bool,reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,reflect.Uintptr,reflect.Float32,reflect.Float64,reflect.Complex64,reflect.Complex128,reflect.Func,reflect.String,reflect.UnsafePointer:
        return realType,false,nil
    case reflect.Ptr:
        sdElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(ptr elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.PtrTo(sdElType),hasStunt,nil
    case reflect.Interface:
        _,has:=cbs[TypeName(realType.String())]; if !has { return realType,false,nil }
        return _STUNT_TYPE,true,nil
    case reflect.Array:
        sdElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(array elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.ArrayOf(realType.Len(),sdElType),hasStunt,nil
    case reflect.Slice:
        sdElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(slice elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.SliceOf(sdElType),hasStunt,nil
    case reflect.Struct:
        // There are some pretty severe limitations of runtime struct type generation.
        // In particular, you can't creates structs with unexported fields.
        // Fortunately, this is usually OK for our use case.
        // I don't try to overcome these limitations -- I just allow StructOf() to panic.
        var sdFields []reflect.StructField; hasStunt:=false
        for i:=0;i<realType.NumField();i++ {
            sdField:=realType.Field(i)
            sdFieldType,hasD,e:=stuntdoubleType(sdField.Type,cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(struct field) error: %v : %v",sdField.Name,e) }
            hasStunt=hasStunt||hasD
            sdField.Type=sdFieldType
            sdFields=append(sdFields,sdField)
        }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.StructOf(sdFields),hasStunt,nil
    case reflect.Map:
        sdKeyType,hasDK,e:=stuntdoubleType(realType.Key(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(map key) error: %v",e) }
        sdElType,hasDE,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(slice elem) error: %v",e) }
        if !(hasDK || hasDE) { return realType,false,nil }
        return reflect.MapOf(sdKeyType,sdElType),true,nil
    case reflect.Chan: return nil,false,fmt.Errorf("Chan unmarshal not yet implemented")  // If I implement this, it could open up some interesting design possibilities...
    default: return nil,false,fmt.Errorf("Unsupported Kind: %v",realType.Kind())
    }
}

// stuntdoubleToReal is the inverse of 'stuntdoubleType'.  It transforms a type
// containing StuntDoubles into a real type.  It uses the callbacks in CBMap to
// accomplish this.
func stuntdoubleToReal(sd,real reflect.Value, cbs CBMap) error {
    sdType:=sd.Type(); realType:=real.Type()

    if sdType==_STUNT_TYPE {
        if cb,has:=cbs[TypeName(realType.String())]; has {
            i,e:=cb([]byte(sd.Interface().(StuntDouble))); if e!=nil { return cbErr{e} }
            sd=reflect.ValueOf(i); sdType=sd.Type()
        }
    }

    // Unmarshalers are always implemented on pointer receivers:
    sdPtrType:=reflect.PtrTo(sdType)
    if sdPtrType.Implements(_JSON_UNMARSHALER_TYPE) || sdPtrType.Implements(_TEXT_UNMARSHALER_TYPE) {
        // Don't descend into this type to avoid losing the custom unmarshaling behavior which probably sets unexported fields that we wouldn't be able to access.
        if !real.CanSet() { return errors.New("cannot set 01") }
        if !sdType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(sd)
        return nil
    }

    switch real.Kind() {
    case reflect.Invalid:
        return errors.New("invalid kind")
    case reflect.Bool,reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,reflect.Uintptr,reflect.Float32,reflect.Float64,reflect.Complex64,reflect.Complex128,reflect.Func,reflect.String,reflect.UnsafePointer:
        if !real.CanSet() { return errors.New("cannot set 02") }
        if !sdType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(sd)
        return nil
    case reflect.Ptr:
        if sd.Kind()!=reflect.Ptr { return errors.New("Incompatible stuntdouble and real kinds") }
        if sd.IsNil() {
            if real.IsNil() { return nil }
            if !real.CanSet() { return errors.New("cannot set 03") }
            real.Set(sd)
            return nil
        }
        if real.IsNil() {
            if !real.CanSet() { return errors.New("cannot set 04") }
            real.Set(reflect.New(real.Type().Elem()))
        }
        return stuntdoubleToReal(sd.Elem(),real.Elem(),cbs)
    case reflect.Interface:
        if !real.CanSet() { return errors.New("cannot set 05") }
        if !sdType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(sd)
        return nil
    case reflect.Array:
        if sd.Kind()!=reflect.Array && sd.Kind()!=reflect.Slice { return errors.New("Incompatible stuntdouble and real kinds") }
        rlen:=real.Len()
        if sd.Len()!=rlen { return errors.New("unequal array lengths") }
        for i:=0;i<rlen;i++ {
            e:=stuntdoubleToReal(sd.Index(i),real.Index(i),cbs); if e!=nil { return fmtErr("array element stuntdoubleToReal error: %v",e) }
        }
        return nil
    case reflect.Slice:
        if sd.Kind()!=reflect.Array && sd.Kind()!=reflect.Slice { return errors.New("Incompatible stuntdouble and real kinds") }
        dlen:=sd.Len()
        s:=reflect.MakeSlice(realType,dlen,dlen)
        for i:=0;i<dlen;i++ {
            e:=stuntdoubleToReal(sd.Index(i),s.Index(i),cbs); if e!=nil { return fmtErr("slice element stuntdoubleToReal error: %v",e) }
        }
        if !real.CanSet() { return errors.New("cannot set 06") }
        real.Set(s)
        return nil
    case reflect.Struct:
        if sd.Kind()!=reflect.Struct { return errors.New("Incompatible stuntdouble and real kinds") }
        rnf:=realType.NumField()
        if sdType.NumField()!=rnf { return errors.New("unequal struct NumFields") }
        for i:=0;i<rnf;i++ {
            rf:=realType.Field(i); df:=sdType.Field(i)
            if rf.Name!=df.Name { return errors.New("unequal struct field names") }
            e:=stuntdoubleToReal(sd.Field(i),real.Field(i),cbs); if e!=nil { return fmtErr("struct field stuntdoubleToReal error: %v",e) }
        }
        return nil
    case reflect.Map:
        if sd.Kind()!=reflect.Map { return errors.New("Incompatible stuntdouble and real kinds") }
        rkeyType:=realType.Key(); rvalType:=realType.Elem()
        m:=reflect.MakeMapWithSize(realType,sd.Len())

        // More efficient way to do it in Go 1.12:
        // iter:=sd.MapRange()
        // for iter.Next() {
        //     dk:=iter.Key(); dv:=iter.Value()

        keys:=sd.MapKeys()
        for _,dk:=range keys {
            dv:=sd.MapIndex(dk)
            rk:=reflect.New(rkeyType).Elem(); rv:=reflect.New(rvalType).Elem()
            e:=stuntdoubleToReal(dk,rk,cbs); if e!=nil { return fmtErr("map key stuntdoubleToReal error: %v",e) }
            e=stuntdoubleToReal(dv,rv,cbs);  if e!=nil { return fmtErr("map val stuntdoubleToReal error: %v",e) }
            m.SetMapIndex(rk,rv)
        }
        if !real.CanSet() { return errors.New("cannot set 07") }
        real.Set(m)
        return nil
    case reflect.Chan: return fmt.Errorf("Chan unmarshal not yet implemented")  // If I implement this, it could open up some interesting design possibilities...
    default: return fmt.Errorf("Unsupported Kind: %v",real.Kind())
    }
}

