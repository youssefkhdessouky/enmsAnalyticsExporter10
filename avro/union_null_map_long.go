// Code generated by github.com/actgardner/gogen-avro/v10. DO NOT EDIT.
/*
 * SOURCE:
 *     stream_data_record_message_schema.avsc
 */
package avro

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/actgardner/gogen-avro/v10/compiler"
	"github.com/actgardner/gogen-avro/v10/vm"
	"github.com/actgardner/gogen-avro/v10/vm/types"
)

type UnionNullMapLongTypeEnum int

const (
	UnionNullMapLongTypeEnumMapLong UnionNullMapLongTypeEnum = 1
)

type UnionNullMapLong struct {
	Null      *types.NullVal
	MapLong   map[string]int64
	UnionType UnionNullMapLongTypeEnum
}

func writeUnionNullMapLong(r *UnionNullMapLong, w io.Writer) error {

	if r == nil {
		err := vm.WriteLong(0, w)
		return err
	}

	err := vm.WriteLong(int64(r.UnionType), w)
	if err != nil {
		return err
	}
	switch r.UnionType {
	case UnionNullMapLongTypeEnumMapLong:
		return writeMapLong(r.MapLong, w)
	}
	return fmt.Errorf("invalid value for *UnionNullMapLong")
}

func NewUnionNullMapLong() *UnionNullMapLong {
	return &UnionNullMapLong{}
}

func (r *UnionNullMapLong) Serialize(w io.Writer) error {
	return writeUnionNullMapLong(r, w)
}

func DeserializeUnionNullMapLong(r io.Reader) (*UnionNullMapLong, error) {
	t := NewUnionNullMapLong()
	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, t)

	if err != nil {
		return t, err
	}
	return t, err
}

func DeserializeUnionNullMapLongFromSchema(r io.Reader, schema string) (*UnionNullMapLong, error) {
	t := NewUnionNullMapLong()
	deser, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, t)

	if err != nil {
		return t, err
	}
	return t, err
}

func (r *UnionNullMapLong) Schema() string {
	return "[\"null\",{\"type\":\"map\",\"values\":\"long\"}]"
}

func (_ *UnionNullMapLong) SetBoolean(v bool)   { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetInt(v int32)      { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetFloat(v float32)  { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetDouble(v float64) { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetBytes(v []byte)   { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetString(v string)  { panic("Unsupported operation") }

func (r *UnionNullMapLong) SetLong(v int64) {

	r.UnionType = (UnionNullMapLongTypeEnum)(v)
}

func (r *UnionNullMapLong) Get(i int) types.Field {

	switch i {
	case 0:
		return r.Null
	case 1:
		r.MapLong = make(map[string]int64)
		return &MapLongWrapper{Target: (&r.MapLong)}
	}
	panic("Unknown field index")
}
func (_ *UnionNullMapLong) NullField(i int)                  { panic("Unsupported operation") }
func (_ *UnionNullMapLong) HintSize(i int)                   { panic("Unsupported operation") }
func (_ *UnionNullMapLong) SetDefault(i int)                 { panic("Unsupported operation") }
func (_ *UnionNullMapLong) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ *UnionNullMapLong) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ *UnionNullMapLong) Finalize()                        {}

func (r *UnionNullMapLong) MarshalJSON() ([]byte, error) {

	if r == nil {
		return []byte("null"), nil
	}

	switch r.UnionType {
	case UnionNullMapLongTypeEnumMapLong:
		return json.Marshal(map[string]interface{}{"map": r.MapLong})
	}
	return nil, fmt.Errorf("invalid value for *UnionNullMapLong")
}

func (r *UnionNullMapLong) UnmarshalJSON(data []byte) error {

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if len(fields) > 1 {
		return fmt.Errorf("more than one type supplied for union")
	}
	if value, ok := fields["map"]; ok {
		r.UnionType = 1
		return json.Unmarshal([]byte(value), &r.MapLong)
	}
	return fmt.Errorf("invalid value for *UnionNullMapLong")
}
