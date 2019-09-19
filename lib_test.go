package jsonface

import (
    "testing"
    "fmt"
    "strings"
    "reflect"
)

type I interface { F() }
type J interface { G() }

type IImpl string
func (me IImpl) F() {}

var cbs=CBMap{ "I":func(bs []byte)(interface{},error){ return `(`+IImpl(bs)+`)`,nil } }

func TestStuntDouble(t *testing.T) {
    d,h,e:=stuntdoubleType(reflect.TypeOf(int32(0)),cbs); if fmt.Sprint(d,h,e)!="int32 false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var i I
    d,h,e=stuntdoubleType(reflect.TypeOf(i),cbs); if e==nil || !strings.Contains(e.Error(),"nil realType") { panic(e) }
    d,h,e=stuntdoubleType(reflect.ValueOf(&i).Elem().Type(),cbs); if fmt.Sprint(d,h,e)!="jsonface.StuntDouble true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var j J
    d,h,e=stuntdoubleType(reflect.ValueOf(&j).Elem().Type(),cbs); if fmt.Sprint(d,h,e)!="jsonface.J false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var is []I
    d,h,e=stuntdoubleType(reflect.TypeOf(is),cbs); if fmt.Sprint(d,h,e)!="[]jsonface.StuntDouble true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var js []J
    d,h,e=stuntdoubleType(reflect.TypeOf(js),cbs); if fmt.Sprint(d,h,e)!="[]jsonface.J false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var ia [10]I
    d,h,e=stuntdoubleType(reflect.TypeOf(ia),cbs); if fmt.Sprint(d,h,e)!="[10]jsonface.StuntDouble true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var ja [10]J
    d,h,e=stuntdoubleType(reflect.TypeOf(ja),cbs); if fmt.Sprint(d,h,e)!="[10]jsonface.J false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var it struct { I I; S string; F float64; B []byte }
    d,h,e=stuntdoubleType(reflect.TypeOf(it),cbs); if fmt.Sprint(d,h,e)!="struct { I jsonface.StuntDouble; S string; F float64; B []uint8 } true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jt struct { J J; S string; F float64; B []byte }
    d,h,e=stuntdoubleType(reflect.TypeOf(jt),cbs); if fmt.Sprint(d,h,e)!="struct { J jsonface.J; S string; F float64; B []uint8 } false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var its []struct { I I; S string; F float64; B []byte }
    d,h,e=stuntdoubleType(reflect.TypeOf(its),cbs); if fmt.Sprint(d,h,e)!="[]struct { I jsonface.StuntDouble; S string; F float64; B []uint8 } true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jts []struct { J J; S string; F float64; B []byte }
    d,h,e=stuntdoubleType(reflect.TypeOf(jts),cbs); if fmt.Sprint(d,h,e)!="[]struct { J jsonface.J; S string; F float64; B []uint8 } false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var im1 map[string]I
    d,h,e=stuntdoubleType(reflect.TypeOf(im1),cbs); if fmt.Sprint(d,h,e)!="map[string]jsonface.StuntDouble true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jm1 map[string]J
    d,h,e=stuntdoubleType(reflect.TypeOf(jm1),cbs); if fmt.Sprint(d,h,e)!="map[string]jsonface.J false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var im2 map[I]string
    d,h,e=stuntdoubleType(reflect.TypeOf(im2),cbs); if fmt.Sprint(d,h,e)!="map[jsonface.StuntDouble]string true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jm2 map[J]string
    d,h,e=stuntdoubleType(reflect.TypeOf(jm2),cbs); if fmt.Sprint(d,h,e)!="map[jsonface.J]string false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var im3 map[string][]I
    d,h,e=stuntdoubleType(reflect.TypeOf(im3),cbs); if fmt.Sprint(d,h,e)!="map[string][]jsonface.StuntDouble true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jm3 map[string][]J
    d,h,e=stuntdoubleType(reflect.TypeOf(jm3),cbs); if fmt.Sprint(d,h,e)!="map[string][]jsonface.J false <nil>" { panic(fmt.Sprint(d,h,e)) }

    var im4 map[string]struct{ I I }
    d,h,e=stuntdoubleType(reflect.TypeOf(im4),cbs); if fmt.Sprint(d,h,e)!="map[string]struct { I jsonface.StuntDouble } true <nil>" { panic(fmt.Sprint(d,h,e)) }

    var jm4 map[string]struct{ J J }
    d,h,e=stuntdoubleType(reflect.TypeOf(jm4),cbs); if fmt.Sprint(d,h,e)!="map[string]struct { J jsonface.J } false <nil>" { panic(fmt.Sprint(d,h,e)) }
}

func TestLib(t *testing.T) {
    var i I
    e:=Unmarshal([]byte(`"cb-success"`),&i,cbs); if fmt.Sprint(i,e)!=`("cb-success")<nil>` { panic(fmt.Sprint(i,e)) }

    var is []I
    e=Unmarshal([]byte(`[0,1.1,"2",[3]]`),&is,cbs); if fmt.Sprint(is,e)!=`[(0) (1.1) ("2") ([3])] <nil>` { panic(fmt.Sprint(is,e)) }

    var ia [4]I
    e=Unmarshal([]byte(`[0,1.1,"2",[3]]`),&ia,cbs); if fmt.Sprint(ia,e)!=`[(0) (1.1) ("2") ([3])] <nil>` { panic(fmt.Sprint(ia,e)) }

    var st struct {
        I I
        S string
    }
    e=Unmarshal([]byte(`{"I":123, "S":"hi"}`),&st,cbs); if fmt.Sprint(st,e)!=`{(123) hi} <nil>` { panic(fmt.Sprint(st,e)) }

    var im map[string]I
    e=Unmarshal([]byte(`{"a":123}`),&im,cbs); if fmt.Sprint(im,e)!=`map[a:(123)] <nil>` { panic(fmt.Sprint(im,e)) }
}

