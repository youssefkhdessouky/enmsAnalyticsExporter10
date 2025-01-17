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

type SqlAttributeBindingDataDto struct {
	TableName *UnionNullString `json:"tableName"`

	Query *UnionNullString `json:"query"`
}

const SqlAttributeBindingDataDtoAvroCRC64Fingerprint = "8\a\xb7h\xa3\xb4@]"

func NewSqlAttributeBindingDataDto() SqlAttributeBindingDataDto {
	r := SqlAttributeBindingDataDto{}
	r.TableName = nil
	r.Query = nil
	return r
}

func DeserializeSqlAttributeBindingDataDto(r io.Reader) (SqlAttributeBindingDataDto, error) {
	t := NewSqlAttributeBindingDataDto()
	deser, err := compiler.CompileSchemaBytes([]byte(t.Schema()), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func DeserializeSqlAttributeBindingDataDtoFromSchema(r io.Reader, schema string) (SqlAttributeBindingDataDto, error) {
	t := NewSqlAttributeBindingDataDto()

	deser, err := compiler.CompileSchemaBytes([]byte(schema), []byte(t.Schema()))
	if err != nil {
		return t, err
	}

	err = vm.Eval(r, deser, &t)
	return t, err
}

func writeSqlAttributeBindingDataDto(r SqlAttributeBindingDataDto, w io.Writer) error {
	var err error
	err = writeUnionNullString(r.TableName, w)
	if err != nil {
		return err
	}
	err = writeUnionNullString(r.Query, w)
	if err != nil {
		return err
	}
	return err
}

func (r SqlAttributeBindingDataDto) Serialize(w io.Writer) error {
	return writeSqlAttributeBindingDataDto(r, w)
}

func (r SqlAttributeBindingDataDto) Schema() string {
	return "{\"fields\":[{\"default\":null,\"name\":\"tableName\",\"type\":[\"null\",\"string\"]},{\"default\":null,\"name\":\"query\",\"type\":[\"null\",\"string\"]}],\"name\":\"com.eventumsolutions.nms.dto.SqlAttributeBindingDataDto\",\"type\":\"record\"}"
}

func (r SqlAttributeBindingDataDto) SchemaName() string {
	return "com.eventumsolutions.nms.dto.SqlAttributeBindingDataDto"
}

func (_ SqlAttributeBindingDataDto) SetBoolean(v bool)    { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetInt(v int32)       { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetLong(v int64)      { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetFloat(v float32)   { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetDouble(v float64)  { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetBytes(v []byte)    { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetString(v string)   { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) SetUnionElem(v int64) { panic("Unsupported operation") }

func (r *SqlAttributeBindingDataDto) Get(i int) types.Field {
	switch i {
	case 0:
		r.TableName = NewUnionNullString()

		return r.TableName
	case 1:
		r.Query = NewUnionNullString()

		return r.Query
	}
	panic("Unknown field index")
}

func (r *SqlAttributeBindingDataDto) SetDefault(i int) {
	switch i {
	case 0:
		r.TableName = nil
		return
	case 1:
		r.Query = nil
		return
	}
	panic("Unknown field index")
}

func (r *SqlAttributeBindingDataDto) NullField(i int) {
	switch i {
	case 0:
		r.TableName = nil
		return
	case 1:
		r.Query = nil
		return
	}
	panic("Not a nullable field index")
}

func (_ SqlAttributeBindingDataDto) AppendMap(key string) types.Field { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) AppendArray() types.Field         { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) HintSize(int)                     { panic("Unsupported operation") }
func (_ SqlAttributeBindingDataDto) Finalize()                        {}

func (_ SqlAttributeBindingDataDto) AvroCRC64Fingerprint() []byte {
	return []byte(SqlAttributeBindingDataDtoAvroCRC64Fingerprint)
}

func (r SqlAttributeBindingDataDto) MarshalJSON() ([]byte, error) {
	var err error
	output := make(map[string]json.RawMessage)
	output["tableName"], err = json.Marshal(r.TableName)
	if err != nil {
		return nil, err
	}
	output["query"], err = json.Marshal(r.Query)
	if err != nil {
		return nil, err
	}
	return json.Marshal(output)
}

func (r *SqlAttributeBindingDataDto) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	var val json.RawMessage
	val = func() json.RawMessage {
		if v, ok := fields["tableName"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.TableName); err != nil {
			return err
		}
	} else {
		r.TableName = NewUnionNullString()

		r.TableName = nil
	}
	val = func() json.RawMessage {
		if v, ok := fields["query"]; ok {
			return v
		}
		return nil
	}()

	if val != nil {
		if err := json.Unmarshal([]byte(val), &r.Query); err != nil {
			return err
		}
	} else {
		r.Query = NewUnionNullString()

		r.Query = nil
	}
	return nil
}
