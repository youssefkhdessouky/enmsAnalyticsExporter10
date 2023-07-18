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
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avro"
	streamingMessageAvro "github.com/youssefkhdessouky/enmsAnalyticsExporter10/avro"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"math"

	"io"
	"os"
	"strconv"
	"sync"
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

func MarshalMetrics(md pmetric.Metrics) map[string][]*streamingMessageAvro.UnionStringNull {

	data := make(map[string][]*streamingMessageAvro.UnionStringNull)
	rms := md.ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.ScopeMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			metrics := ilm.Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
				appendMetricDescriptons(metric, data)
				appendMetricDataPoints(metric, data)
                fmt.Println("change test")
			}
		}
	}

	return data
}
func appendMetricDataPoints(m pmetric.Metric, data map[string][]*streamingMessageAvro.UnionStringNull) {
	var MetricTypeData string = ""

	switch m.Type() {
	case pmetric.MetricTypeEmpty:
		break
	case pmetric.MetricTypeGauge:

		points := m.Gauge()
		pts := points.DataPoints()
		MetricTypeData += appendNumberDataPoints(pts)

	case pmetric.MetricTypeSum:
		points := m.Sum()
		MetricTypeData += "[ IsMonotonic : " + strconv.FormatBool(points.IsMonotonic()) + ", "
		MetricTypeData += "AggregationTemporality : " + points.AggregationTemporality().String() + ", "

		//data["IsMonotonic"] = append(data["IsMonotonic"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatBool(points.IsMonotonic()),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		//
		//data["AggregationTemporality"] = append(data["AggregationTemporality"], &streamingMessageAvro.UnionStringNull{String: points.AggregationTemporality().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		//u, err := json.Marshal(points.DataPoints())
		//if err != nil {
		//	fmt.Println("Error marshalling " + pmetric.MetricTypeGauge.String())
		//}
		pts := points.DataPoints()


		MetricTypeData += appendNumberDataPoints(pts)

		//appendNumberDataPoints(points.DataPoints(), data)
		MetricTypeData += " ]"

	case pmetric.MetricTypeHistogram:
		points := m.Histogram()
		tmp := &points
		//data["AggregationTemporality"] = append(data["AggregationTemporality"], &streamingMessageAvro.UnionStringNull{String: points.AggregationTemporality().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		MetricTypeData += "[ IsMonotonic : " + "" + ", "
		MetricTypeData += "AggregationTemporality : " + points.AggregationTemporality().String() + ", "

		pts := tmp.DataPoints()

		MetricTypeData += appendHistogramDataPoints(pts)

		MetricTypeData += "] "
		//appendHistogramDataPoints(points.DataPoints(), data)
	case pmetric.MetricTypeExponentialHistogram:
		points := m.ExponentialHistogram()
		MetricTypeData += "[ AggregationTemporality : " + points.AggregationTemporality().String() + ", "
		//data["AggregationTemporality"] = append(data["AggregationTemporality"], &streamingMessageAvro.UnionStringNull{String: points.AggregationTemporality().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		pts := points.DataPoints()
		MetricTypeData += appendExponentialHistogramDataPoints(pts)
		//appendExponentialHistogramDataPoints(points.DataPoints(), data)
		MetricTypeData += " ]"
	case pmetric.MetricTypeSummary:

		points := m.Summary()

		tmp := &points
		pts := tmp.DataPoints()

		MetricTypeData += appendDoubleSummaryDataPoints(pts)

	}

	data["Metric_value"] = append(data["Metric_value"], &streamingMessageAvro.UnionStringNull{String: MetricTypeData,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

}

func appendNumberDataPoints(ps pmetric.NumberDataPointSlice) string {
	pts_data := "{ Datapoints : ["
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)

		pts_data += appendDataPointAttributes(p.Attributes())

		pts_data += "{ AggregationTemporality : " + p.StartTimestamp().String() + " }, "
		//data["AggregationTemporality"] = append(data["AggregationTemporality"], &streamingMessageAvro.UnionStringNull{String: p.StartTimestamp().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ StartTimestamp : " + p.Timestamp().String() + " }, "
		//data["StartTimestamp"] = append(data["StartTimestamp"], &streamingMessageAvro.UnionStringNull{String: p.Timestamp().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		switch p.ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			pts_data += "{ Value : " + strconv.FormatInt(p.IntValue(), 10) + " } "
			//data["Value"] = append(data["Value"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatInt(p.IntValue(), 10),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		case pmetric.NumberDataPointValueTypeDouble:
			pts_data += "{ Value : " + strconv.FormatFloat(p.DoubleValue(), 'f', -1, 64) + " } "
			//data["Value"] = append(data["Value"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.DoubleValue(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

	}
	pts_data += " ] }"
	return pts_data
}

// log attributes
func appendDataPointAttributes(attributes pcommon.Map) string {
	if attributes.Len() == 0 {
		return ""
	}
	attr := "{ attributes : [ "
	attributes.Range(func(k string, v pcommon.Value) bool {
		//data[k] = append(data[k], &streamingMessageAvro.UnionStringNull{String: v.AsString(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		attr += "{ " + k + " : " + v.AsString() + "}, "

		return true
	})

	attr += " ] }"

	return attr
}

func appendDoubleSummaryDataPoints(ps pmetric.SummaryDataPointSlice) string {
	pts_data := "{ Datapoints : ["

	for i := 0; i < ps.Len(); i++ {

		p := ps.At(i)
		pts_data += " { SummaryDataPoints : " + strconv.Itoa(i) + " }, "

		pts_data += appendDataPointAttributes(p.Attributes())

		pts_data += "{ StartTimestamp : " + p.StartTimestamp().AsTime().String() + " }, "

		//data["StartTimestamp"] = append(data["StartTimestamp"], &streamingMessageAvro.UnionStringNull{String: p.StartTimestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Timestamp : " + p.Timestamp().AsTime().String() + " }, "

		//data["Timestamp"] = append(data["Timestamp"], &streamingMessageAvro.UnionStringNull{String: p.Timestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Count : " + strconv.FormatUint(p.Count(), 10) + " }, "
		//data["Count"] = append(data["Count"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatUint(p.Count(), 10),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Sum : " + strconv.FormatFloat(p.Sum(), 'f', -1, 64) + " }, "
		//data["Sum"] = append(data["Sum"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Sum(), 'f', -1, 64),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		quantiles := p.QuantileValues()
		quan_data := "{ QuantileValues : ["
		for i := 0; i < quantiles.Len(); i++ {
			quantile := quantiles.At(i)
			if i+1 != quantiles.Len() {
				quan_data += "{ " + strconv.FormatFloat(quantile.Quantile(), 'f', -1, 64) + " : " + strconv.FormatFloat(quantile.Value(), 'f', -1, 64) + " }, "
			} else {
				quan_data += "{ " + strconv.FormatFloat(quantile.Quantile(), 'f', -1, 64) + " : " + strconv.FormatFloat(quantile.Value(), 'f', -1, 64) + " } "
			}
			//data[strconv.FormatFloat(quantile.Quantile(), 'f', -1, 64)] = append(data[strconv.FormatFloat(quantile.Quantile(), 'f', -1, 64)], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(quantile.Value(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		}
		quan_data += " ] }"
		if i+1 != ps.Len() {
			pts_data += quan_data + " ,"
		} else {
			pts_data += quan_data
		}
	}
	pts_data += " ] }"

	return pts_data
}

func appendExponentialHistogramDataPoints(ps pmetric.ExponentialHistogramDataPointSlice) string {
	pts_data := "{ Datapoints : ["
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)

		pts_data += "{ ExponentialHistogramDataPoints : " + strconv.Itoa(i) + " }, "
		//data["ExponentialHistogramDataPoints"] = append(data["ExponentialHistogramDataPoints"], &streamingMessageAvro.UnionStringNull{String: strconv.Itoa(i),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += appendDataPointAttributes(p.Attributes())

		pts_data += "{ StartTimestamp : " + p.StartTimestamp().AsTime().String() + " }, "
		//data["StartTimestamp"] = append(data["StartTimestamp"], &streamingMessageAvro.UnionStringNull{String: p.StartTimestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Timestamp : " + p.Timestamp().AsTime().String() + " }, "
		//data["Timestamp"] = append(data["Timestamp"], &streamingMessageAvro.UnionStringNull{String: p.Timestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Count : " + strconv.FormatUint(p.Count(), 10) + " }, "
		//data["Count"] = append(data["Count"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatUint(p.Count(), 10),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		if p.HasSum() {
			pts_data += "{ Sum : " + strconv.FormatFloat(p.Sum(), 'f', -1, 64) + " }, "
			//data["Sum"] = append(data["Sum"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Sum(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		} else {
			pts_data += "{ Sum : }, "
			//data["Sum"] = append(data["Sum"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		if p.HasMin() {
			pts_data += "{ Min : " + strconv.FormatFloat(p.Min(), 'f', -1, 64) + " }, "
			//data["Min"] = append(data["Min"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Min(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		} else {
			pts_data += "{ Min : }, "
			//data["Min"] = append(data["Min"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		if p.HasMax() {
			pts_data += "{ Max : " + strconv.FormatFloat(p.Max(), 'f', -1, 64) + " }, "
			//data["Max"] = append(data["Max"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Max(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		} else {
			pts_data += "{ Max : }, "
			//data["Max"] = append(data["Max"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		scale := int(p.Scale())
		factor := math.Ldexp(math.Ln2, -scale)
		// Note: the equation used here, which is
		//   math.Exp(index * factor)
		// reports +Inf as the _lower_ boundary of the bucket nearest
		// infinity, which is incorrect and can be addressed in various
		// ways.  The OTel-Go implementation of this histogram pending
		// in https://github.com/open-telemetry/opentelemetry-go/pull/2393
		// uses a lookup table for the last finite boundary, which can be
		// easily computed using `math/big` (for scales up to 20).

		negB := p.Negative().BucketCounts()
		posB := p.Positive().BucketCounts()
		pts_expo := "{ points : [ "
		for i := 0; i < negB.Len(); i++ {
			pos := negB.Len() - i - 1
			index := p.Negative().Offset() + int32(pos)
			lower := math.Exp(float64(index) * factor)
			upper := math.Exp(float64(index+1) * factor)

			position := "[-" + strconv.FormatFloat(upper, 'f', -1, 64) + ", " + "-" + strconv.FormatFloat(lower, 'f', -1, 64) + "-]"
			count := strconv.FormatUint(negB.At(pos), 10)

			pts_expo += "{" + position + " : " + count + "}"

			//data[position] = append(data[position], &streamingMessageAvro.UnionStringNull{String: count,
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		}

		if p.ZeroCount() != 0 {
			pts_expo += "{ [0,0] : " + strconv.FormatUint(p.ZeroCount(), 10) + " }"
			//data["[0,0]"] = append(data["[0,0]"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatUint(p.ZeroCount(), 10),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		} else {
			pts_expo += "{ [0,0] : }"
			//data["[0,0]"] = append(data["[0,0]"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		for pos := 0; pos < posB.Len(); pos++ {
			index := p.Positive().Offset() + int32(pos)
			lower := math.Exp(float64(index) * factor)
			upper := math.Exp(float64(index+1) * factor)

			position := "[ " + strconv.FormatFloat(lower, 'f', -1, 64) + ", " + "-" + strconv.FormatFloat(upper, 'f', -1, 64) + "]"
			count := strconv.FormatUint(posB.At(pos), 10)

			pts_expo += "{" + position + " : " + count + " }"
			//data[position] = append(data[position], &streamingMessageAvro.UnionStringNull{String: count,
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}
		if i+1 != ps.Len() {
			pts_data += pts_expo + ", "
		} else {
			pts_data += pts_expo
		}
	}
	pts_data += " ] }"

	return pts_data
}

func appendHistogramDataPoints(ps pmetric.HistogramDataPointSlice) string {
	pts_data := "{ Datapoints : ["
	for i := 0; i < ps.Len(); i++ {
		p := ps.At(i)
		pts_data += "{ HistogramDataPoints : " + strconv.Itoa(i) + " }, "
		//data["HistogramDataPoints"] = append(data["HistogramDataPoints"], &streamingMessageAvro.UnionStringNull{String: strconv.Itoa(i),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += appendDataPointAttributes(p.Attributes())

		pts_data += "{ StartTimestamp : " + p.StartTimestamp().AsTime().String() + " }, "
		//data["StartTimestamp"] = append(data["StartTimestamp"], &streamingMessageAvro.UnionStringNull{String: p.StartTimestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		pts_data += "{ Timestamp : " + p.Timestamp().AsTime().String() + " }, "
		//data["Timestamp"] = append(data["Timestamp"], &streamingMessageAvro.UnionStringNull{String: p.Timestamp().AsTime().String(),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		if p.HasSum() {
			pts_data += "{ Sum : " + strconv.FormatFloat(p.Sum(), 'f', -1, 64) + " }, "
			//data["Sum"] = append(data["Sum"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Sum(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		} else {
			pts_data += "{ Sum : }, "
			//data["Sum"] = append(data["Sum"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		if p.HasMin() {
			pts_data += "{ Min : " + strconv.FormatFloat(p.Min(), 'f', -1, 64) + " }, "
			//data["Min"] = append(data["Min"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Min(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		} else {
			pts_data += "{ Min : }, "
			//data["Min"] = append(data["Min"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		if p.HasMax() {
			pts_data += "{ Max : " + strconv.FormatFloat(p.Max(), 'f', -1, 64) + " }, "
			//data["Max"] = append(data["Max"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatFloat(p.Max(), 'f', -1, 64),
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		} else {
			pts_data += "{ Max : }, "
			//data["Max"] = append(data["Max"], &streamingMessageAvro.UnionStringNull{String: "",
			//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
		}

		pts_data += "{ ExplicitBounds : " + arrayFloat64ToString(p.ExplicitBounds().AsRaw()) + " }, "

		//data["ExplicitBounds"] = append(data["ExplicitBounds"], &streamingMessageAvro.UnionStringNull{String: arrayFloat64ToString(p.ExplicitBounds().AsRaw()),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

		pts_data += "{ Buckets : " + arrayUInt64ToString(p.BucketCounts().AsRaw()) + " }"

		if i+1 != ps.Len() {
			pts_data += ", "
		}
		//data["Buckets"] = append(data["Buckets"], &streamingMessageAvro.UnionStringNull{String: arrayUInt64ToString(p.BucketCounts().AsRaw()),
		//	UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

	}
	pts_data += " }"
	return pts_data
}

func arrayUInt64ToString(raw []uint64) string {
	var result string = ""
	for i, value := range raw {
		result += strconv.FormatUint(value, 10)
		if i < len(raw)-1 {
			result += ","
		}
	}
	return result
}
func arrayFloat64ToString(floatArray []float64) string {
	var result string = ""
	for i, value := range floatArray {
		result += strconv.FormatFloat(value, 'f', -1, 64)
		if i < len(floatArray)-1 {
			result += ","
		}
	}
	return result
}
func appendMetricDescriptons(md pmetric.Metric, data map[string][]*streamingMessageAvro.UnionStringNull) {

	name := md.Name()

	unit := md.Unit()

	metricType := md.Type().String()

	data["Name"] = append(data["Name"], &streamingMessageAvro.UnionStringNull{String: name,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

	data["Unit"] = append(data["Unit"], &streamingMessageAvro.UnionStringNull{String: unit,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

	data["Type"] = append(data["Type"], &streamingMessageAvro.UnionStringNull{String: metricType,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

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

	fmt.Errorf("consuming traces <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
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
	columnNames := [3]string{"SpanId", "TraceId", "ParentId"}
	for _, columnName := range columnNames {
		data[columnName] = make([]*streamingMessageAvro.UnionStringNull, 0)
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

func appendToData(span ptrace.Span, data map[string][]*streamingMessageAvro.UnionStringNull) {
	low, high := TraceIDToUInt64Pair(span.TraceID())
	traceIdString := string(low) + "-" + string(high)
	startTime := span.StartTimestamp().String()
	endTime := span.EndTimestamp().String()
	name := span.Name()
	kind := span.Kind().String()
	statusCode := span.Status().Code().String()
	statusMsg := span.Status().Message()
	attributes := appendDataPointAttributes(span.Attributes())
	events := logEvents(span.Events())
	links := logLinks(span.Links())

	data["SpanId"] = append(data["SpanId"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatUint(SpanIDToUInt64(span.SpanID()), 10),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["TraceId"] = append(data["TraceId"], &streamingMessageAvro.UnionStringNull{String: traceIdString,
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

	data["ParentId"] = append(data["ParentId"], &streamingMessageAvro.UnionStringNull{String: strconv.FormatUint(SpanIDToUInt64(span.ParentSpanID()), 10),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["StartTime"] = append(data["StartTime"], &streamingMessageAvro.UnionStringNull{String: string(startTime),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["EndTime"] = append(data["EndTime"], &streamingMessageAvro.UnionStringNull{String: string(endTime),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["Name"] = append(data["Name"], &streamingMessageAvro.UnionStringNull{String: string(name),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["Kind"] = append(data["Kind"], &streamingMessageAvro.UnionStringNull{String: string(kind),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["StatusCode"] = append(data["StatusCode"], &streamingMessageAvro.UnionStringNull{String: string(statusCode),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["StatusMessage"] = append(data["StatusMessage"], &streamingMessageAvro.UnionStringNull{String: string(statusMsg),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["Events"] = append(data["Events"], &streamingMessageAvro.UnionStringNull{String: string(events),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["Links"] = append(data["Links"], &streamingMessageAvro.UnionStringNull{String: string(links),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})
	data["Attributes"] = append(data["Attributes"], &streamingMessageAvro.UnionStringNull{String: string(attributes),
		UnionType: streamingMessageAvro.UnionStringNullTypeEnumString})

}

func (e *enmsAnalyticsExporter) ConsumeMetrics(_ context.Context, md pmetric.Metrics) error {
	buf, err := e.metricsMarshaler.MarshalMetrics(md)
	if err != nil {
		return err
	}
	buf = e.compressor(buf)
	topic := "records"

	fmt.Errorf("consuming Traces <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
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

	value.Data.MapArrayUnionStringNull = MarshalMetrics(md)

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

func (e *enmsAnalyticsExporter) ConsumeLogs(_ context.Context, ld plog.Logs) error {
	buf, err := e.logsMarshaler.MarshalLogs(ld)
	if err != nil {
		return err
	}
	buf = e.compressor(buf)
	return e.exporter(e, buf)
}
func logInstrumentationScope(il pcommon.InstrumentationScope) string {

	instrumentationScope := "{ InstrumentationScope : [ "

	instrumentationScope += "{ Name : " + il.Name() + ", "
	instrumentationScope += "Version : " + il.Version() + ", "

	instrumentationScope += appendDataPointAttributes(il.Attributes())
	instrumentationScope += " ] }"

	return instrumentationScope
}
func logLinks(sl ptrace.SpanLinkSlice) string {
	if sl.Len() == 0 {
		return ""
	}
	spanLinks := "{ spanLinks : [ "

	for i := 0; i < sl.Len(); i++ {
		l := sl.At(i)

		spanLinkNum := "{ SpanLink #" + strconv.Itoa(i) + " : "
		spanLinks += spanLinkNum
		spanLinks += "TraceId : " + l.TraceID().String() + ", "
		spanLinks += "SpanID : " + l.SpanID().String() + ", "
		spanLinks += "TraceState : " + l.TraceState().AsRaw() + ", "
		spanLinks += "DroppedAttributesCount : " + strconv.FormatUint(uint64(l.DroppedAttributesCount()), 10) + ", "

		if l.Attributes().Len() == 0 {
			continue
		}
		spanLinks += appendDataPointAttributes(l.Attributes())
		spanLinks += " }, "

	}
	spanLinks += " ] }"
	return spanLinks

}
func logEvents(se ptrace.SpanEventSlice) string {
	if se.Len() == 0 {
		return ""
	}
	spanEvents := "{ SpanEvents : [ "

	for i := 0; i < se.Len(); i++ {
		e := se.At(i)

		spanEventNum := "{ SpanEvent #" + strconv.Itoa(i) + " : "
		spanEvents += spanEventNum
		spanEvents += "Name : " + e.Name() + ", "
		spanEvents += "Timestamp : " + e.Timestamp().String() + ", "
		spanEvents += "DroppedAttributesCount : " + strconv.FormatUint(uint64(e.DroppedAttributesCount()), 10) + ", "

		if e.Attributes().Len() == 0 {
			continue
		}
		spanEvents += appendDataPointAttributes(e.Attributes())
		spanEvents += " }, "

	}
	spanEvents += " ] }"
	return spanEvents
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
