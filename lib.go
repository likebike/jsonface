package jsonface

import (
    "fmt"
    "errors"
    "reflect"
    "encoding"
    stdjson "encoding/json"
    "sync"
)

type RawMessage=stdjson.RawMessage

type CBMap map[TypeName]CB  // By mapping from TypeName to CB, it's possible to have name conflicts.  If I encounter that problem in real-life, I might change the mapping to reflect.Type-to-CB.
type TypeName string
type CB func([]byte) (interface{},error)

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
var _JSON_UNMARSHALER_TYPE=reflect.TypeOf((*stdjson.Unmarshaler)(nil)).Elem()
var _TEXT_UNMARSHALER_TYPE=reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

var globalCBs=struct {
    sync.RWMutex
    m CBMap
}{sync.RWMutex{},CBMap{}}

func AddGlobalCB(name TypeName, cb CB) {
    globalCBs.Lock(); defer globalCBs.Unlock()
    if _,has:=globalCBs.m[name]; has { panic(errors.New("CB already defined")) }
    globalCBs.m[name]=cb
}

func Marshal(x interface{}) ([]byte,error) { return stdjson.Marshal(x) }

func StdUnmarshal(bs []byte, destPtr interface{}) error { return stdjson.Unmarshal(bs,destPtr) }

func GlobalUnmarshal(bs []byte, destPtr interface{}) error {
    globalCBs.RLock(); defer globalCBs.RUnlock()
    return Unmarshal(bs,destPtr,globalCBs.m)
}

func Unmarshal(bs []byte, destPtr interface{}, cbs CBMap) error {
    destPtrV:=reflect.ValueOf(destPtr)
    if !destPtrV.IsValid() { return errors.New("invalid destPtr") }
    if destPtrV.Kind()!=reflect.Ptr { return errors.New("destPtr is not a pointer") }
    if destPtrV.IsNil() { return errors.New("nil destPtr") }
    destType:=destPtrV.Elem().Type(); if destType==nil { return errors.New("nil destType") }
    dumType,hasStunt,e:=stuntdoubleType(destType,cbs); if e!=nil { return fmt.Errorf("stuntdoubleType error: %v",e) }
    if !hasStunt { return stdjson.Unmarshal(bs,destPtr) }  // If no stunt was used, just fallback to standard behavior.
    dumPtrV:=reflect.New(dumType)
    if !dumPtrV.CanInterface() { return errors.New("cannot dumPtrV.Interface()") }
    e=stdjson.Unmarshal(bs,dumPtrV.Interface()); if e!=nil { return fmt.Errorf("json.Unmarshal error: %v",e) }
    e=stuntdoubleToReal(dumPtrV,destPtrV,cbs); if e!=nil { return unwrapCBErr(fmtErr("stuntdoubleToReal error: %v",e)) }
    return nil
}

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
        dumElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(ptr elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.PtrTo(dumElType),hasStunt,nil
    case reflect.Interface:
        _,has:=cbs[TypeName(realType.Name())]; if !has { return realType,false,nil }
        return _STUNT_TYPE,true,nil
    case reflect.Array:
        dumElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(array elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.ArrayOf(realType.Len(),dumElType),hasStunt,nil
    case reflect.Slice:
        dumElType,hasStunt,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(slice elem) error: %v",e) }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.SliceOf(dumElType),hasStunt,nil
    case reflect.Struct:
        // There are some pretty severe limitations of runtime struct type generation.
        // In particular, you can't creates structs with unexported fields.
        // Fortunately, this is usually OK for our use case.
        // I don't try to overcome these limitations -- I just allow StructOf() to panic.
        var dumFields []reflect.StructField; hasStunt:=false
        for i:=0;i<realType.NumField();i++ {
            dumField:=realType.Field(i)
            dumFieldType,hasD,e:=stuntdoubleType(dumField.Type,cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(struct field) error: %v : %v",dumField.Name,e) }
            hasStunt=hasStunt||hasD
            dumField.Type=dumFieldType
            dumFields=append(dumFields,dumField)
        }
        if !hasStunt { return realType,hasStunt,nil }
        return reflect.StructOf(dumFields),hasStunt,nil
    case reflect.Map:
        dumKeyType,hasDK,e:=stuntdoubleType(realType.Key(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(map key) error: %v",e) }
        dumElType,hasDE,e:=stuntdoubleType(realType.Elem(),cbs); if e!=nil { return nil,false,fmt.Errorf("stuntdoubleType(slice elem) error: %v",e) }
        if !(hasDK || hasDE) { return realType,false,nil }
        return reflect.MapOf(dumKeyType,dumElType),true,nil
    case reflect.Chan: return nil,false,fmt.Errorf("Chan unmarshal not yet implemented")  // If I implement this, it could open up some interesting design possibilities...
    default: return nil,false,fmt.Errorf("Unsupported Kind: %v",realType.Kind())
    }
}

func stuntdoubleToReal(dum,real reflect.Value, cbs CBMap) error {
    dumType:=dum.Type(); realType:=real.Type()

    if dumType==_STUNT_TYPE {
        if cb,has:=cbs[TypeName(realType.Name())]; has {
            i,e:=cb([]byte(dum.Interface().(StuntDouble))); if e!=nil { return cbErr{e} }
            dum=reflect.ValueOf(i); dumType=dum.Type()
        }
    }

    // Unmarshalers are always implemented on pointer receivers:
    dumPtrType:=reflect.PtrTo(dumType)
    if dumPtrType.Implements(_JSON_UNMARSHALER_TYPE) || dumPtrType.Implements(_TEXT_UNMARSHALER_TYPE) {
        // Don't descend into this type to avoid losing the custom unmarshaling behavior which probably sets unexported fields that we wouldn't be able to access.
        if !real.CanSet() { return errors.New("cannot set 01") }
        if !dumType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(dum)
        return nil
    }

    switch real.Kind() {
    case reflect.Invalid:
        return errors.New("invalid kind")
    case reflect.Bool,reflect.Int,reflect.Int8,reflect.Int16,reflect.Int32,reflect.Int64,reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64,reflect.Uintptr,reflect.Float32,reflect.Float64,reflect.Complex64,reflect.Complex128,reflect.Func,reflect.String,reflect.UnsafePointer:
        if !real.CanSet() { return errors.New("cannot set 02") }
        if !dumType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(dum)
        return nil
    case reflect.Ptr:
        if dum.Kind()!=reflect.Ptr { return errors.New("Incompatible stuntdouble and real kinds") }
        if dum.IsNil() {
            if real.IsNil() { return nil }
            if !real.CanSet() { return errors.New("cannot set 03") }
            real.Set(dum)
            return nil
        }
        if real.IsNil() {
            if !real.CanSet() { return errors.New("cannot set 04") }
            real.Set(reflect.New(real.Type().Elem()))
        }
        return stuntdoubleToReal(dum.Elem(),real.Elem(),cbs)
    case reflect.Interface:
        if !real.CanSet() { return errors.New("cannot set 05") }
        if !dumType.AssignableTo(realType) { return fmt.Errorf("cb result not assignable") }
        real.Set(dum)
        return nil
    case reflect.Array:
        if dum.Kind()!=reflect.Array && dum.Kind()!=reflect.Slice { return errors.New("Incompatible stuntdouble and real kinds") }
        rlen:=real.Len()
        if dum.Len()!=rlen { return errors.New("unequal array lengths") }
        for i:=0;i<rlen;i++ {
            e:=stuntdoubleToReal(dum.Index(i),real.Index(i),cbs); if e!=nil { return fmtErr("array element stuntdoubleToReal error: %v",e) }
        }
        return nil
    case reflect.Slice:
        if dum.Kind()!=reflect.Array && dum.Kind()!=reflect.Slice { return errors.New("Incompatible stuntdouble and real kinds") }
        dlen:=dum.Len()
        s:=reflect.MakeSlice(realType,dlen,dlen)
        for i:=0;i<dlen;i++ {
            e:=stuntdoubleToReal(dum.Index(i),s.Index(i),cbs); if e!=nil { return fmtErr("slice element stuntdoubleToReal error: %v",e) }
        }
        if !real.CanSet() { return errors.New("cannot set 06") }
        real.Set(s)
        return nil
    case reflect.Struct:
        if dum.Kind()!=reflect.Struct { return errors.New("Incompatible stuntdouble and real kinds") }
        rnf:=realType.NumField()
        if dumType.NumField()!=rnf { return errors.New("unequal struct NumFields") }
        for i:=0;i<rnf;i++ {
            rf:=realType.Field(i); df:=dumType.Field(i)
            if rf.Name!=df.Name { return errors.New("unequal struct field names") }
            e:=stuntdoubleToReal(dum.Field(i),real.Field(i),cbs); if e!=nil { return fmtErr("struct field stuntdoubleToReal error: %v",e) }
        }
        return nil
    case reflect.Map:
        if dum.Kind()!=reflect.Map { return errors.New("Incompatible stuntdouble and real kinds") }
        rkeyType:=realType.Key(); rvalType:=realType.Elem()
        m:=reflect.MakeMapWithSize(realType,dum.Len())

        // More efficient way to do it in Go 1.12:
        // iter:=dum.MapRange()
        // for iter.Next() {
        //     dk:=iter.Key(); dv:=iter.Value()

        keys:=dum.MapKeys()
        for _,dk:=range keys {
            dv:=dum.MapIndex(dk)
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

