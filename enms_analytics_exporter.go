// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enmsAnalyticsExporter // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"

import (
	"context"
	"encoding/binary"
	"io"
	"sync"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	streamingMessageAvro "avro"
	"os"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"strconv"
)

// Marshaler configuration used for marhsaling Protobuf
var tracesMarshalers = map[string]ptrace.Marshaler{
	formatTypeJSON:  &ptrace.JSONMarshaler{},
	formatTypeProto: &ptrace.ProtoMarshaler{},
}
var metricsMarshalers = map[string]pmetric.Marshaler{
	formatTypeJSON:  &pmetric.JSONMarshaler{},
	formatTypeProto: &pmetric.ProtoMarshaler{},
}
var logsMarshalers = map[string]plog.Marshaler{
	formatTypeJSON:  &plog.JSONMarshaler{},
	formatTypeProto: &plog.ProtoMarshaler{},
}

// exportFunc defines how to export encoded telemetry data.
type exportFunc func(e *enmsAnalyticsExporter, buf []byte) error

// enmsAnalyticsExporter is the implementation of enmsAnalyticsExporter that sends telemetry data to enms analytics layer

type enmsAnalyticsExporter struct {
	path  string
	file  io.WriteCloser
	mutex sync.Mutex

	tracesMarshaler  ptrace.Marshaler
	metricsMarshaler pmetric.Marshaler
	logsMarshaler    plog.Marshaler

	compression string
	compressor  compressFunc

	formatType string
	exporter   exportFunc
}

func (e *enmsAnalyticsExporter) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (e *enmsAnalyticsExporter) ConsumeTraces(_ context.Context, td ptrace.Traces) error {

	buf, err := e.tracesMarshaler.MarshalTraces(td)
	if err != nil {
		return err
	}
	buf = e.compressor(buf)
	topic := "records"

	fmt.Println("consuming traces <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "192.168.45.34:9092"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created Producer %v\n", p)

	client, err := schemaregistry.NewClient(schemaregistry.NewConfig("http://192.168.45.34:8088"))

	if err != nil {
		fmt.Printf("Failed to create schema registry client: %s\n", err)
		os.Exit(1)
	}

	ser, err := avro.NewSpecificSerializer(client, serde.ValueSerde, avro.NewSerializerConfig())

	if err != nil {
		fmt.Printf("Failed to create serializer: %s\n", err)
		os.Exit(1)
	}

	// Optional delivery channel, if not specified the Producer object's
	// .Events channel is used.
	deliveryChan := make(chan kafka.Event)

	value := streamingMessageAvro.StreamDataRecordMessage{}
	value.Data = &streamingMessageAvro.UnionNullMapArrayUnionStringNull{}
	value.Data.UnionType = streamingMessageAvro.UnionNullMapArrayUnionStringNullTypeEnumMapArrayUnionStringNull
	value.Data.MapArrayUnionStringNull = make(map[string][]*streamingMessageAvro.UnionStringNull)
	// columnNames:= [3]string {"SpanId", "TraceId", "ParentId"}
	// for _, columnName := range columnNames {
	// 	value.Data.MapArrayUnionStringNull[columnName] = make([]*streamingMessageAvro.UnionStringNull,0)
	// }
	// spanIds := [4]string {"1", "2", "3", "4"}
	// traceIds := [4]string {"1", "2", "1", "2"}
	// parentIds := [4]string {"","1", "1", "2"}
	// for _, spanId := range spanIds {
	// 	value.Data.MapArrayUnionStringNull["SpanId"] = append(value.Data.MapArrayUnionStringNull["SpanId"], &streamingMessageAvro.UnionStringNull{String: spanId,
	// 		 UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})
	// }
	// for _, traceId := range traceIds {
	// 	value.Data.MapArrayUnionStringNull["TraceId"] = append(value.Data.MapArrayUnionStringNull["TraceId"], &streamingMessageAvro.UnionStringNull{String:traceId,
	// 		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})
	// }
	// for _, parentId := range parentIds {
	// 	value.Data.MapArrayUnionStringNull["ParentId"] = append(value.Data.MapArrayUnionStringNull["ParentId"], &streamingMessageAvro.UnionStringNull{String:parentId,
	// 		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})
	// }
	


	resourceSpans := td.ResourceSpans()

	if resourceSpans.Len() == 0 {
		return nil
	}

	for i := 0; i < resourceSpans.Len(); i++ {
		rs := resourceSpans.At(i)
		value.Data.MapArrayUnionStringNull = resourceSpansToJaegerProto(rs)
	}
	payload, err := ser.Serialize(topic, &value)
	if err != nil {
		fmt.Printf("Failed to serialize payload: %s\n", err)
		os.Exit(1)
	}

	err = p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          payload,
		Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
	}, deliveryChan)
	if err != nil {
		fmt.Printf("Produce failed: %v\n", err)
		os.Exit(1)
	}

	eD := <-deliveryChan
	m := eD.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		fmt.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	} else {
		fmt.Printf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
	close(deliveryChan)

	return e.exporter(e, buf)
}

func resourceSpansToJaegerProto(rs ptrace.ResourceSpans) map[string][]*streamingMessageAvro.UnionStringNull {
	resource := rs.Resource()
	ilss := rs.ScopeSpans()

	if resource.Attributes().Len() == 0 && ilss.Len() == 0 {
		return nil
	}
	
	data := make(map[string][]*streamingMessageAvro.UnionStringNull)

	if ilss.Len() == 0 {
		return data
	}

	// Approximate the number of the spans as the number of the spans in the first
	// instrumentation library info.
	columnNames:= [3]string {"SpanId", "TraceId", "ParentId"}
	for _, columnName := range columnNames {
		data[columnName] = make([]*streamingMessageAvro.UnionStringNull,0)
	}
	for i := 0; i < ilss.Len(); i++ {
		ils := ilss.At(i)
		spans := ils.Spans()
		for j := 0; j < spans.Len(); j++ {
			span := spans.At(j)
			appendToData(span, data)
	
		}
	}


	return data
}

func TraceIDToUInt64Pair(traceID pcommon.TraceID) (uint64, uint64) {
	return binary.BigEndian.Uint64(traceID[:8]), binary.BigEndian.Uint64(traceID[8:])
}

func SpanIDToUInt64(spanID pcommon.SpanID) uint64 {
	return binary.BigEndian.Uint64(spanID[:])
}

func appendToData(span ptrace.Span, data map[string][]*streamingMessageAvro.UnionStringNull)  {
	low, high := TraceIDToUInt64Pair(span.TraceID())
	traceIdString := string(low)+"-" + string(high)
	startTime := span.StartTimestamp().String()
	data["SpanId"] = append(data["SpanId"], &streamingMessageAvro.UnionStringNull{String:strconv.FormatUint(SpanIDToUInt64(span.SpanID()), 10) ,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})
	data["TraceId"] = append(data["TraceId"], &streamingMessageAvro.UnionStringNull{String:traceIdString,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})

	data["ParentId"] = append(data["ParentId"], &streamingMessageAvro.UnionStringNull{String:strconv.FormatUint(SpanIDToUInt64(span.ParentSpanID()), 10),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})
	data["StartTime"] = append(data["StartTime"], &streamingMessageAvro.UnionStringNull{String:string(startTime),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString,})}	

func (e *enmsAnalyticsExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	buf, err := e.metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	buf = e.compressor(buf)
	return e.exporter(e, buf)
}

func (e *enmsAnalyticsExporter) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	buf, err := e.logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	buf = e.compressor(buf)
	return e.exporter(e, buf)
}

func exportMessageAsLine(e *enmsAnalyticsExporter, buf []byte) error {
	// Ensure only one write operation happens at a time.
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if _, err := e.file.Write(buf); err != nil {
		return err
	}
	if _, err := io.WriteString(e.file, "\n"); err != nil {
		return err
	}
	return nil
}

func exportMessageAsBuffer(e *enmsAnalyticsExporter, buf []byte) error {
	// Ensure only one write operation happens at a time.
	e.mutex.Lock()
	defer e.mutex.Unlock()
	// write the size of each message before writing the message itself.  https://developers.google.com/protocol-buffers/docs/techniques
	// each encoded object is preceded by 4 bytes (an unsigned 32 bit integer)
	data := make([]byte, 4, 4+len(buf))
	binary.BigEndian.PutUint32(data, uint32(len(buf)))
	data = append(data, buf...)
	if err := binary.Write(e.file, binary.BigEndian, data); err != nil {
		return err
	}
	return nil
}

func (e *enmsAnalyticsExporter) Start(context.Context, component.Host) error {
	return nil
}

// Shutdown stops the exporter and is invoked during shutdown.
func (e *enmsAnalyticsExporter) Shutdown(context.Context) error {
	return e.file.Close()
}

func buildExportFunc(cfg *Config) func(e *enmsAnalyticsExporter, buf []byte) error {
	if cfg.FormatType == formatTypeProto {
		return exportMessageAsBuffer
	}
	// if the data format is JSON and needs to be compressed, telemetry data can't be written to file in JSON format.
	if cfg.FormatType == formatTypeJSON && cfg.Compression != "" {
		return exportMessageAsBuffer
	}
	return exportMessageAsLine
}
