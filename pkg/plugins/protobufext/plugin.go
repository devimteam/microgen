package protobufext

import (
	"reflect"
	"time"

	"github.com/devimteam/microgen/pkg/plugins"
)

func init() {
	plugins.RegisterProtobufTypeBinding(protoext{})
}

const (
	googleProtobuf             = "google.protobuf."
	googleProtobufStringValue  = googleProtobuf + "StringValue"
	googleProtobufBoolValue    = googleProtobuf + "BoolValue"
	googleProtobufInt64Value   = googleProtobuf + "Int64Value"
	googleProtobufUInt64Value  = googleProtobuf + "UInt64Value"
	googleProtobufInt32Value   = googleProtobuf + "Int32Value"
	googleProtobufUInt32Value  = googleProtobuf + "UInt32Value"
	googleProtobufFloat64Value = googleProtobuf + "DoubleValue"
	googleProtobufFloat32Value = googleProtobuf + "FloatValue"
	googleProtobufTimestamp    = googleProtobuf + "Timestamp"

	importGoogleProtobuf          = "google/protobuf/"
	importGoogleProtobufWrappers  = importGoogleProtobuf + "wrappers.proto"
	importGoogleProtobufTimestamp = importGoogleProtobuf + "timestamp.proto"
)

type protoext struct{}

func (protoext) ProtobufType(origType reflect.Type) (pbType reflect.Type, ok bool) {
	panic("implement me")
}

func (protoext) MarshalLayout(origType reflect.Type) (marshalLayout string, ok bool) {
	panic("implement me")
}

func (protoext) UnmarshalLayout(origType reflect.Type) (unmarshalLayout string, ok bool) {
	panic("implement me")
}

func (protoext) ProtoBinding(origType reflect.Type) (fieldType string, requiredImport *string, ok bool) {
	switch origType {
	case stringPType:
		return googleProtobufStringValue, sp(importGoogleProtobufWrappers), true
	case boolPType:
		return googleProtobufBoolValue, sp(importGoogleProtobufWrappers), true
	case intPType:
		return googleProtobufInt64Value, sp(importGoogleProtobufWrappers), true
	case int32PType:
		return googleProtobufInt32Value, sp(importGoogleProtobufWrappers), true
	case int64PType:
		return googleProtobufInt64Value, sp(importGoogleProtobufWrappers), true
	case uintPType:
		return googleProtobufUInt64Value, sp(importGoogleProtobufWrappers), true
	case uint32PType:
		return googleProtobufUInt32Value, sp(importGoogleProtobufWrappers), true
	case uint64PType:
		return googleProtobufUInt64Value, sp(importGoogleProtobufWrappers), true
	case float64PType:
		return googleProtobufFloat64Value, sp(importGoogleProtobufWrappers), true
	case float32PType:
		return googleProtobufFloat32Value, sp(importGoogleProtobufWrappers), true
	case timeType:
		return googleProtobufTimestamp, sp(importGoogleProtobufTimestamp), true
	case timePType:
		return googleProtobufTimestamp, sp(importGoogleProtobufTimestamp), true
	default:
		return "", nil, false
	}
}

func sp(s string) *string {
	return &s
}

var (
	stringPType  = reflect.TypeOf(new(*string)).Elem()
	boolPType    = reflect.TypeOf(new(*bool)).Elem()
	intPType     = reflect.TypeOf(new(int)).Elem()
	int32PType   = reflect.TypeOf(new(int32)).Elem()
	int64PType   = reflect.TypeOf(new(int64)).Elem()
	uintPType    = reflect.TypeOf(new(uint)).Elem()
	uint32PType  = reflect.TypeOf(new(uint32)).Elem()
	uint64PType  = reflect.TypeOf(new(uint64)).Elem()
	timeType     = reflect.TypeOf(new(time.Time)).Elem()
	timePType    = reflect.TypeOf(new(*time.Time)).Elem()
	float32PType = reflect.TypeOf(new(*float32)).Elem()
	float64PType = reflect.TypeOf(new(*float64)).Elem()
)
