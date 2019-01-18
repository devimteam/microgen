package protobufext

import (
	"github.com/golang/protobuf/ptypes/wrappers"
)

func String_ToProtobuf(str string) *wrappers.StringValue {
	if str == "" {
		return nil
	}
	return &wrappers.StringValue{Value: str}
}

func String_FromProtobuf(value *wrappers.StringValue) string {
	if value == nil {
		return ""
	}
	t := value.Value
	return t
}

func P_String_ToProtobuf(str *string) *wrappers.StringValue {
	if str == nil {
		return nil
	}
	return String_ToProtobuf(*str)
}

func P_String_FromProtobuf(value *wrappers.StringValue) *string {
	if value == nil {
		return nil
	}
	ret := String_FromProtobuf(value)
	return &ret
}

func P_UInt32_ToProtobuf(d *uint32) *wrappers.UInt32Value {
	if d == nil {
		return nil
	}
	return &wrappers.UInt32Value{
		Value: *d,
	}
}

func P_UInt32_FromProtobuf(d *wrappers.UInt32Value) *uint32 {
	if d == nil {
		return nil
	}
	t := d.Value
	return &t
}

func P_UInt64_ToProtobuf(d *uint64) *wrappers.UInt64Value {
	if d == nil {
		return nil
	}
	return &wrappers.UInt64Value{
		Value: *d,
	}
}

func P_UInt64_FromProtobuf(d *wrappers.UInt64Value) *uint64 {
	if d == nil {
		return nil
	}
	t := d.Value
	return &t
}

func P_Int64_ToProtobuf(d *int64) *wrappers.Int64Value {
	if d == nil {
		return nil
	}
	return &wrappers.Int64Value{
		Value: *d,
	}
}

func P_Int64_FromProtobuf(d *wrappers.Int64Value) *int64 {
	if d == nil {
		return nil
	}
	t := d.Value
	return &t
}

func P_Int32_ToProtobuf(d *int32) *wrappers.Int32Value {
	if d == nil {
		return nil
	}
	return &wrappers.Int32Value{
		Value: *d,
	}
}

func P_Int32_FromProtobuf(d *wrappers.Int32Value) *int32 {
	if d == nil {
		return nil
	}
	t := d.Value
	return &t
}

func P_Bool_ToProtobuf(d *bool) *wrappers.BoolValue {
	if d == nil {
		return nil
	}
	return &wrappers.BoolValue{
		Value: *d,
	}
}

func P_Bool_FromProtobuf(d *wrappers.BoolValue) *bool {
	if d == nil {
		return nil
	}
	t := d.Value
	return &t
}
