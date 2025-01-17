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

var _ = fmt.Printf

type DateTimeColumnFormatDTO struct {
	ColumnName *UnionNullString `json:"columnName"`

	Format *UnionNullString `json:"format"`
}

const DateTimeColumnFormatDTOAvroCRC64Fingerprint = "\xfd\xc19 f6&V"

func NewDateTimeColumnFormatDTO() DateTimeColumnFormatDTO {
	r := DateTimeColumnFormatDTO{}
	r.ColumnName = nil
	r.Format = nil
	return r
}

func DeserializeDateTimeColumnFormatDTO(r io.Reader) (DateTimeColumnFormatDTO, error) {
	t := NewDateTimeColumnFormatDTO()
	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func DeserializeDateTimeColumnFormatDTOFromSchema(r io.Reader, schema string) (DateTimeColumnFormatDTO, error) {
	t := NewDateTimeColumnFormatDTO()

	deser, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func writeDateTimeColumnFormatDTO(r DateTimeColumnFormatDTO, w io.Writer) error {
	var err error
	err = writeUnionNullString(r.ColumnName, w)
	if err != nil {
		return err
	}
	err = writeUnionNullString(r.Format, w)
	if err != nil {
		return err
	}
	return err
}

func (r DateTimeColumnFormatDTO) Serialize(w io.Writer) error {
	return writeDateTimeColumnFormatDTO(r, w)
}

func (r DateTimeColumnFormatDTO) Schema() string {
	return "{\"fields\":[{\"default\":null,\"name\":\"columnName\",\"type\":[\"null\",\"string\"]},{\"default\":null,\"name\":\"format\",\"type\":[\"null\",\"string\"]}],\"name\":\"com.eventumsolutions.nms.kafka.messages.streaming.DateTimeColumnFormatDTO\",\"type\":\"record\"}"
}

func (r DateTimeColumnFormatDTO) SchemaName() string {
	return "com.eventumsolutions.nms.kafka.messages.streaming.DateTimeColumnFormatDTO"
}

func (_ DateTimeColumnFormatDTO) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetInt(v int32)       { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetLong(v int64)      { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetString(v string)   { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) SetUnionElem(v int64) { panic("Unsupported operation") }

func (r *DateTimeColumnFormatDTO) Get(i int) types.Field {
	switch i {
	case 0:
		r.ColumnName = NewUnionNullString()

		return r.ColumnName
	case 1:
		r.Format = NewUnionNullString()

		return r.Format
	}
	panic("Unknown field index")
}

func (r *DateTimeColumnFormatDTO) SetDefault(i int) {
	switch i {
	case 0:
		r.ColumnName = nil
		return
	case 1:
		r.Format = nil
		return
	}
	panic("Unknown field index")
}

func (r *DateTimeColumnFormatDTO) NullField(i int) {
	switch i {
	case 0:
		r.ColumnName = nil
		return
	case 1:
		r.Format = nil
		return
	}
	panic("Not a nullable field index")
}

func (_ DateTimeColumnFormatDTO) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) HintSize(int)                     { panic("Unsupported operation") }
func (_ DateTimeColumnFormatDTO) Finalize()                        {}

func (_ DateTimeColumnFormatDTO) AvroCRC64Fingerprint() []byte {
	return []byte(DateTimeColumnFormatDTOAvroCRC64Fingerprint)
}

func (r DateTimeColumnFormatDTO) MarshalJSON() ([]byte, error) {
	var err error
	output := make(map[string]json.RawMessage)
	output["columnName"], err = json.Marshal(r.ColumnName)
	if err != nil {
		return nil, err
	}
	output["format"], err = json.Marshal(r.Format)
	if err != nil {
		return nil, err
	}
	return json.Marshal(output)
}

func (r *DateTimeColumnFormatDTO) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var val json.RawMessage
	val = func() json.RawMessage {
		if v, ok := fields["columnName"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.ColumnName); err != nil {
			return err
		}
	} else {
		r.ColumnName = NewUnionNullString()

		r.ColumnName = nil
	}
	val = func() json.RawMessage {
		if v, ok := fields["format"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.Format); err != nil {
			return err
		}
	} else {
		r.Format = NewUnionNullString()

		r.Format = nil
	}
	return nil
}
