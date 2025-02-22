package jsonvalue

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"unsafe"
)

func testStructConv(t *testing.T) {
	cv("export to string", func() { testExportString(t) })
	cv("export to int", func() { testExportInt(t) })
	cv("export to float", func() { testExportFloat(t) })
	cv("export to bool", func() { testExportBool(t) })
	cv("misc import", func() { testImport(t) })
	cv("test structconv.go Import", func() { testStructConv_Import(t) })

	cv("test Issue 19", func() { testImportBugIssue19(t) })
	cv("test Issue 22", func() { testImportBugIssue22(t) })

	cv("test miscellaneous anonymous situations", func() { testImportMiscAnonymous(t) })
}

func testExportString(t *testing.T) {
	const S = "Hello, jsonvalue!"
	v := NewString(S)

	str := ""
	err := v.Export(str)
	so(err, isErr)

	err = v.Export(&str)
	so(err, isNil)

	so(str, eq, S)

	bol := true
	err = v.Export(&bol)
	so(err, isErr)

	v = &V{}
	err = v.Export(nil)
	so(err, isErr)
}

func testExportInt(t *testing.T) {
	const positive = 123454321
	const negative = -987656789

	n1 := NewInt(positive)

	var i int
	var u uint
	var i32 int32
	var u32 uint32

	err := n1.Export(&i)
	so(err, isNil)
	so(i, eq, positive)

	err = n1.Export(&u)
	so(err, isNil)
	so(u, eq, positive)

	err = n1.Export(&i32)
	so(err, isNil)
	so(i32, eq, positive)

	err = n1.Export(&u32)
	so(err, isNil)
	so(u32, eq, positive)

	// --------

	n2 := NewInt(negative)

	err = n2.Export(&i)
	so(err, isNil)
	so(i, eq, negative)

	err = n2.Export(&i32)
	so(err, isNil)
	so(i32, eq, negative)

	// --------

	bol := true
	err = n1.Export(&bol)
	so(err, isErr)
}

func testExportFloat(t *testing.T) {
	const F = 12345.4321

	n := NewFloat64(F)

	var f32 float32
	var f64 float64

	err := n.Export(&f32)
	so(err, isNil)
	so(f32, eq, F)

	err = n.Export(&f64)
	so(err, isNil)
	so(f64, eq, F)

	// --------

	bol := true
	err = n.Export(&bol)
	so(err, isErr)
}

func testExportBool(t *testing.T) {
	v := NewBool(true)
	b := false

	err := v.Export(b)
	so(err, isErr)

	err = v.Export(&b)
	so(err, isNil)

	so(b, isTrue)

	str := ""
	err = v.Export(&str)
	so(err, isErr)
}

func testImport(t *testing.T) {
	cv("integers", func() {

		params := []any{
			int(1),
			uint(2),
			int8(3),
			uint8(4),
			int16(5),
			uint16(6),
			int32(7),
			uint32(8),
			int64(9),
			uint64(10),
		}

		for i, p := range params {
			v, err := Import(p)
			so(err, isNil)
			so(v.ValueType(), eq, Number)
			so(v.Int(), eq, i+1)
		}
	})

	cv("string", func() {
		s := "hello"
		v, err := Import(s)
		so(err, isNil)
		so(v.ValueType(), eq, String)
		so(v.String(), eq, s)
	})

	cv("object", func() {
		type thing struct {
			String string `json:"str"`
		}
		th := thing{
			String: "world",
		}

		v, err := Import(&th)
		so(err, isNil)
		so(v.ValueType(), eq, Object)

		s, err := v.GetString("str")
		so(err, isNil)
		so(s, eq, th.String)
	})

	cv("float", func() {
		m := map[string]interface{}{
			"32": float32(0.023),
			"64": float64(0.023),
		}

		v, err := Import(m)
		so(err, isNil)

		s := v.MustMarshalString(OptDefaultStringSequence())
		so(s, eq, `{"32":0.023,"64":0.023}`)
	})

	cv("error", func() {
		f := func() bool {
			return false
		}
		v, err := Import(f)
		so(err, isErr)
		so(v, notNil)
		so(v.ValueType(), eq, NotExist)
	})
}

func testStructConv_Import(t *testing.T) {
	cv("[]byte, json.RawMessage", func() { testStructConv_Import_RawAndBytes(t) })
	cv("uintptr", func() { testStructConv_Import_StrangeButSupportedTypes(t) })
	cv("invalid types", func() { testStructConv_Import_InvalidTypes(t) })
	cv("general types", func() { testStructConv_Import_NormalTypes(t) })
	cv("array and slice", func() { testStructConv_Import_ArrayAndSlice(t) })
}

func testStructConv_Import_RawAndBytes(t *testing.T) {
	cv("json.RawMessage", func() {
		msg := "Hello, raw message!"
		st := struct {
			Raw json.RawMessage `json:"raw"`
		}{
			Raw: []byte(fmt.Sprintf(`{"message":"%s"}`, msg)),
		}

		j, err := Import(&st)
		so(err, isNil)
		so(j.Len(), eq, 1)

		got, err := j.GetString("raw", "message")
		so(err, isNil)
		so(got, eq, msg)

		// t.Logf("%v", j)
	})

	cv("json.RawMessage error", func() {
		msg := "Hello, raw message!"
		st := struct {
			Raw json.RawMessage `json:"raw"`
		}{
			Raw: []byte(fmt.Sprintf(`{"message":"%s"`, msg)),
		}

		j, err := Import(&st)
		so(err, isErr)
		so(j, notNil)
	})

	cv("[]byte", func() {
		b := []byte{1, 2, 3, 4}
		st := struct {
			Bytes []byte `json:"bytes"`
		}{
			Bytes: b,
		}

		j, err := Import(&st)
		so(err, isNil)

		gotS, err := j.GetString("bytes")
		so(err, isNil)

		gotB, err := j.GetBytes("bytes")
		so(err, isNil)

		gotStoB, err := base64.StdEncoding.DecodeString(gotS)
		so(err, isNil)

		so(bytes.Equal(gotB, b), isTrue)
		so(bytes.Equal(gotStoB, b), isTrue)

		// t.Logf("%v", j)
	})
}

func testStructConv_Import_StrangeButSupportedTypes(t *testing.T) {
	cv("uintptr", func() {
		st := struct {
			Ptr uintptr
		}{
			Ptr: 1234,
		}

		b, err := json.Marshal(&st)
		t.Logf("Got bytes: '%s'", b)
		so(err, isNil)

		j, err := Import(&st)
		so(err, isNil)

		bb := j.MustMarshal()
		so(bytes.Equal(b, bb), isTrue)
	})

	cv("map[uintptr]xxx", func() {
		m := map[uintptr]int{
			1: 2,
			2: 3,
		}

		j, err := Import(m)
		so(err, isNil)
		so(j.IsObject(), isTrue)
		so(j.MustGet("1").Uint(), eq, m[1])
		so(j.MustGet("2").Uint(), eq, m[2])
	})

	cv("map[int]xxx", func() {
		m := map[int]int{
			1: 2,
			2: 3,
		}

		j, err := Import(m)
		so(err, isNil)
		so(j.IsObject(), isTrue)
		so(j.MustGet("1").Int(), eq, m[1])
		so(j.MustGet("2").Int(), eq, m[2])
	})
}

func testStructConv_Import_InvalidTypes(t *testing.T) {
	cv("complex", func() {
		st := struct {
			C complex128
		}{
			C: complex(1, 2),
		}

		_, err := json.Marshal(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)

		_, err = Import(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)
	})

	cv("chan", func() {
		st := struct {
			Ch chan struct{}
		}{
			Ch: make(chan struct{}),
		}

		//lint:ignore SA1026 intend to do this to check error in uni-test
		_, err := json.Marshal(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)

		_, err = Import(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)
	})

	cv("unsafe.Pointer", func() {
		st := struct {
			Ptr unsafe.Pointer
		}{
			Ptr: nil,
		}

		_, err := json.Marshal(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)

		_, err = Import(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)
	})

	cv("func", func() {
		st := struct {
			Func func()
		}{
			Func: func() { panic("Hey!") },
		}

		//lint:ignore SA1026 intend to do this to check error in uni-test
		_, err := json.Marshal(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)

		_, err = Import(&st)
		so(err, isErr)
		t.Logf("expect error: %v", err)
	})

	cv("not a struct", func() {
		ch := make(chan error)
		defer close(ch)

		j, err := Import(&ch)
		so(err, isErr)
		so(j, notNil)

		j, err = Import(ch)
		so(err, isErr)
		so(j, notNil)
	})

	cv("map[float64]xxx", func() {
		m := map[float64]int{
			1: 2,
			2: 3,
		}

		j, err := Import(m)
		// panic(j.MustMarshalString())
		so(err, isErr)
		so(j, notNil)
	})

	cv("illegal type in slice", func() {
		arr := []any{
			1, complex(1, 2),
		}
		j, err := Import(arr)
		so(err, isErr)
		so(j, notNil)

		arr = []any{
			1, []any{complex(1, 2)},
		}
		j, err = Import(arr)
		so(err, isErr)
		so(j, notNil)
	})

	cv("illegal type in map", func() {
		m := map[string]any{
			"complex": complex(1, 2),
		}
		j, err := Import(m)
		so(err, isErr)
		so(j, notNil)

		m = map[string]any{
			"obj": map[string]any{
				"complex": complex(1, 2),
			},
		}
		j, err = Import(m)
		so(err, isErr)
		so(j, notNil)
	})
}

func testStructConv_Import_NormalTypes(t *testing.T) {
	cv("bool", func() {
		st := struct {
			True   bool `json:"true"`
			False  bool `json:"false"`
			Empty  bool `json:",omitempty"`
			String bool `json:",string"`
		}{
			True:   true,
			False:  false,
			Empty:  false,
			String: true,
		}

		b, err := json.Marshal(&st)
		t.Logf("Got bytes: '%s'", b)
		so(err, isNil)

		j, err := Import(&st)
		t.Logf("Got bytes: '%s'", j.MustMarshalString())
		so(err, isNil)

		boo, err := j.GetBool("true")
		so(err, isNil)
		so(boo, isTrue)

		boo, err = j.GetBool("false")
		so(err, isNil)
		so(boo, isFalse)

		_, err = j.GetBool("Empty")
		so(err, isErr)

		boo, err = j.GetBool("String")
		so(err, isErr)
		so(boo, isTrue)
	})

	cv("number", func() {
		st := struct {
			Int     int32   `json:"int,string"`
			Uint    uint64  `json:"uint,string"`
			Float32 float32 `json:"float32,string"`
			Float64 float64 `json:"float64,string"`
		}{
			Int:     -100,
			Uint:    10000,
			Float32: 123.125,
			Float64: 123.125,
		}

		j, err := Import(&st)
		so(err, isNil)

		s, err := j.GetString("int")
		so(err, isNil)
		so(s, eq, strconv.Itoa(int(st.Int)))

		s, err = j.GetString("uint")
		so(err, isNil)
		so(s, eq, strconv.FormatUint(uint64(st.Uint), 10))

		s, err = j.GetString("float32")
		so(err, isNil)
		so(s, eq, strconv.FormatFloat(float64(st.Float32), 'f', -1, 32))

		s, err = j.GetString("float64")
		so(err, isNil)
		so(s, eq, strconv.FormatFloat(float64(st.Float64), 'f', -1, 64))
	})
}

func testStructConv_Import_ArrayAndSlice(t *testing.T) {
	cv("slice", func() {
		st := []struct {
			S string `json:"string"`
			I int    `json:"int"`
		}{
			{
				S: "Hello, 01",
				I: 1,
			}, {
				S: "Hello, 02",
				I: 2,
			},
		}

		j, err := Import(&st)
		so(err, isNil)

		// t.Logf("%s", j.MustMarshalString())

		so(j.IsArray(), isTrue)
		so(j.Len(), eq, 2)

		for i := range j.ForRangeArr() {
			s, err := j.GetString(i, "string")
			so(err, isNil)
			so(s, eq, st[i].S)

			n, err := j.GetInt(i, "int")
			so(err, isNil)
			so(n, eq, st[i].I)
		}
	})

	cv("array", func() {
		arr := [6]rune{'你', '好', 'J', 'S', 'O', 'N'}

		j, err := Import(&arr)
		so(err, isNil)
		so(j.IsArray(), isTrue)

		so(j.Len(), eq, len(arr))
		for i, r := range arr {
			child, err := j.Get(i)
			so(err, isNil)
			so(child.IsNumber(), isTrue)
			so(child.Uint(), eq, r)
		}
	})

	cv("map[string]any", func() {
		m := map[string]any{
			"uint":   uint8(255),
			"float":  float32(-0.25),
			"string": "Hello, any",
			"bool":   true,
			"struct": struct{}{},
		}

		j, err := Import(m)
		so(err, isNil)
		so(j.IsObject(), isTrue)

		// t.Log(j.MustMarshalString())

		so(j.MustGet("uint").Uint(), eq, m["uint"])
		so(j.MustGet("float").Float64(), eq, m["float"])
		so(j.MustGet("string").String(), eq, m["string"])
		so(j.MustGet("bool").Bool(), eq, m["bool"])

		st, err := j.Get("struct")
		so(err, isNil)
		so(st.IsObject(), isTrue)
		so(st.Len(), isZero)
	})

	cv("anonymous struct", func() {
		type inner struct {
			F float64
		}
		st := struct {
			inner
			I int
		}{}

		st.F = 0.25
		st.I = 1024

		j, err := Import(&st)
		so(err, isNil)

		f, err := j.GetFloat32("F")
		so(err, isNil)
		so(f, eq, st.F)

		i, err := j.GetFloat32("I")
		so(err, isNil)
		so(i, eq, st.I)
	})

	cv("illegal anonymous struct", func() {
		type inner struct {
			Ch chan error
		}
		st := struct {
			inner
			I int
		}{}

		j, err := Import(&st)
		so(err, isErr)
		so(j, notNil)
	})

	cv("misc omitempty", func() {
		st := struct {
			I    int             `json:",omitempty"`
			U    uint            `json:",omitempty"`
			S    string          `json:",omitempty"`
			F    float64         `json:",omitempty"`
			A    []int           `json:",omitempty"`
			B    bool            `json:",omitempty"`
			By   []byte          `json:",omitempty"`
			Rw   json.RawMessage `json:",omitempty"`
			St   *struct{}       `json:",omitempty"`
			M    map[int]int     `json:",omitempty"`
			Null *int            `json:"null"`

			priv   int
			Ignore bool `json:"-"`
		}{}

		b, err := json.Marshal(&st)
		so(err, isNil)
		s := string(b)

		j, err := Import(&st)
		so(err, isNil)
		so(j.MustMarshalString(), eq, s)

		st.M = map[int]int{}
		st.A = []int{}

		j, err = Import(&st)
		so(err, isNil)
		so(j.MustMarshalString(), eq, s)

		t.Log(s)
	})

	cv("not omitempty", func() {
		st := struct {
			A    []int
			B    bool
			By   []byte
			St   *struct{}
			M    map[int]int
			Null *int
		}{
			A:  []int{},
			By: []byte{},
			M:  map[int]int{},
		}

		j, err := Import(&st)
		so(err, isNil)

		so(j.MustGet("A").IsArray(), isTrue)
		so(j.MustGet("B").IsBoolean(), isTrue)
		so(j.MustGet("By").IsString(), isTrue)
		so(j.MustGet("St").IsNull(), isTrue)
		so(j.MustGet("M").IsObject(), isTrue)
		so(j.MustGet("Null").IsNull(), isTrue)
	})

	cv("stringfied value", func() {
		type st struct {
			Int  int  `json:"int,string"`
			Bool bool `json:"bool,string"`
		}
		s := st{
			Int:  100,
			Bool: true,
		}
		b, _ := json.Marshal(s)
		t.Logf("json marshal result: %s", string(b))
		v, err := Import(s)
		so(err, isNil)
		so(v.ValueType(), eq, Object)
		so(v.MustGet("int").ValueType(), eq, String)
		so(v.MustGet("int").Int(), eq, s.Int)
		so(v.MustGet("bool").ValueType(), eq, String)
		so(v.MustGet("bool").Bool(), eq, s.Bool)
	})
}

func testImportBugIssue19(t *testing.T) {
	type req struct {
		IDs []uint64 `json:"ids,omitempty"`
	}
	type data struct {
		Req *req `json:"req,omitempty"`
	}

	d := data{}
	d.Req = &req{}
	d.Req.IDs = []uint64{0}

	j, err := Import(d)
	so(err, isNil)

	s := j.MustMarshalString()
	so(s, eq, `{"req":{"ids":[0]}}`)
}

func testImportBugIssue22(t *testing.T) {
	type inner struct {
		Name string
	}
	type outer struct {
		*inner
		Age int
	}
	person := &outer{}
	person.inner = &inner{}
	person.Name = "Andrew"

	b, _ := json.Marshal(person)
	t.Logf("encoding/json: %s", b)

	s := New(person).MustMarshalString(OptSetSequence())
	so(s, eq, string(b))
}

func testImportMiscAnonymous(t *testing.T) {
	cv("struct pointer", func() { testImportMiscAnonymousStructPtrInStruct(t) })
	cv("empty struct pointer", func() { testImportMiscEmptyAnonymousStructPtrInStruct(t) })
	cv("exportable basic types", func() { testImportMiscAnonymousExportableBasicTypeInStruct(t) })
	cv("exportable basic types with tags", func() { testImportMiscAnonymousExportableBasicTypeInStructWithTags(t) })
	cv("un-exportable basic types", func() { testImportMiscAnonymousPrivateBasicTypeInStruct(t) })
	cv("slice", func() { testImportMiscAnonymousSliceInStruct(t) })
	cv("array", func() { testImportMiscAnonymousArrayInStruct(t) })
	cv("slice pointer", func() { testImportMiscAnonymousSlicePtrInStruct(t) })
	cv("invalid types", func() { testImportMiscAnonymousInvalidTypes(t) })
}

func testImportMiscAnonymousStructPtrInStruct(t *testing.T) {
	type inner struct {
		Name string
	}
	type outer struct {
		*inner
		Age int
	}

	person := &outer{}
	person.inner = &inner{}
	person.Name = "Andrew"
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":"Andrew","Age":20}`)
}

func testImportMiscEmptyAnonymousStructPtrInStruct(t *testing.T) {
	type inner struct {
		Name string
	}
	type outer struct {
		*inner
		Age int
	}

	person := &outer{}
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString()

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Age":20}`)
}

func testImportMiscAnonymousExportableBasicTypeInStruct(t *testing.T) {
	type Name string
	type Age int
	type Gender string

	type outer struct {
		Name
		Age
		Gender `json:"-"`
	}

	person := &outer{}
	person.Name = "Andrew"
	person.Age = 20
	person.Gender = "male"

	//lint:ignore SA9005 because the lint is error
	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":"Andrew","Age":20}`)
}

func testImportMiscAnonymousExportableBasicTypeInStructWithTags(t *testing.T) {
	type Name string
	type Age int

	type outer struct {
		Name `json:"n"`
		Age  `json:"a"`
	}

	person := &outer{}
	person.Name = "Andrew"
	person.Age = 20

	//lint:ignore SA9005 because the lint is error
	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"n":"Andrew","a":20}`)
}

func testImportMiscAnonymousPrivateBasicTypeInStruct(t *testing.T) {
	type name string

	type outer struct {
		name `json:"n"`
		Age  float64 `json:"a"`
	}

	person := &outer{}
	person.name = "Andrew"
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"a":20}`)
}

func testImportMiscAnonymousSliceInStruct(t *testing.T) {
	type Name []string
	type nickname []string

	type outer struct {
		Name
		nickname
		Age int
	}

	person := &outer{}
	person.Name = []string{"Andrew", "M", "C"}
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":["Andrew","M","C"],"Age":20}`)

	person.nickname = nickname{"AMC"}

	b, _ = json.Marshal(person)
	v, err = Import(person)
	s = v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":["Andrew","M","C"],"Age":20}`)
}

func testImportMiscAnonymousArrayInStruct(t *testing.T) {
	type Name [3]string

	type outer struct {
		Name
		Age int
	}

	person := &outer{}
	person.Name[0] = "Andrew"
	person.Name[1] = "C"
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":["Andrew","C",""],"Age":20}`)
}

func testImportMiscAnonymousSlicePtrInStruct(t *testing.T) {
	type Name []string

	type outer struct {
		*Name

		Age int
	}

	person := &outer{}
	person.Age = 20

	b, _ := json.Marshal(person)
	v, err := Import(person)
	s := v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":null,"Age":20}`)

	n := Name{"Andrew", "M", "C"}
	person.Name = &n

	b, _ = json.Marshal(person)
	v, err = Import(person)
	s = v.MustMarshalString(OptSetSequence())

	so(err, isNil)
	so(s, eq, string(b))
	so(s, eq, `{"Name":["Andrew","M","C"],"Age":20}`)
}

func testImportMiscAnonymousInvalidTypes(t *testing.T) {
	type Inner chan int
	type outer struct {
		Inner
		Age int
	}

	data := &outer{}
	//lint:ignore SA1026 intent to test this
	_, err := json.Marshal(data)
	so(err, isErr)

	_, err = Import(data)
	so(err, isErr)
}
