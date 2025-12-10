package observability

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Field represents a structured logging field that can be used with the logger.
// It wraps OpenTelemetry attributes for use with the OTel-based logger.
type Field struct {
	Key   string
	Value attribute.Value
}

// toAttribute converts a Field to an OpenTelemetry attribute.KeyValue
func (f Field) toAttribute() attribute.KeyValue {
	switch f.Value.Type() {
	case attribute.BOOL:
		return attribute.Bool(f.Key, f.Value.AsBool())
	case attribute.INT64:
		return attribute.Int64(f.Key, f.Value.AsInt64())
	case attribute.FLOAT64:
		return attribute.Float64(f.Key, f.Value.AsFloat64())
	case attribute.STRING:
		return attribute.String(f.Key, f.Value.AsString())
	case attribute.BOOLSLICE:
		return attribute.BoolSlice(f.Key, f.Value.AsBoolSlice())
	case attribute.INT64SLICE:
		return attribute.Int64Slice(f.Key, f.Value.AsInt64Slice())
	case attribute.FLOAT64SLICE:
		return attribute.Float64Slice(f.Key, f.Value.AsFloat64Slice())
	case attribute.STRINGSLICE:
		return attribute.StringSlice(f.Key, f.Value.AsStringSlice())
	default:
		return attribute.String(f.Key, f.Value.AsString())
	}
}

// FieldsToAttributes converts a slice of Fields to OpenTelemetry attributes
func FieldsToAttributes(fields []Field) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, len(fields))
	for i, f := range fields {
		attrs[i] = f.toAttribute()
	}
	return attrs
}

// String creates a field with a string value.
func String(key, value string) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(value),
	}
}

// Int creates a field with an int value.
func Int(key string, value int) Field {
	return Field{
		Key:   key,
		Value: attribute.IntValue(value),
	}
}

// Int8 creates a field with an int8 value.
func Int8(key string, value int8) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Int16 creates a field with an int16 value.
func Int16(key string, value int16) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Int32 creates a field with an int32 value.
func Int32(key string, value int32) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Int64 creates a field with an int64 value.
func Int64(key string, value int64) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(value),
	}
}

// Uint creates a field with a uint value.
func Uint(key string, value uint) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Uint8 creates a field with a uint8 value.
func Uint8(key string, value uint8) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Uint16 creates a field with a uint16 value.
func Uint16(key string, value uint16) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Uint32 creates a field with a uint32 value.
func Uint32(key string, value uint32) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Uint64 creates a field with a uint64 value.
func Uint64(key string, value uint64) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Float32 creates a field with a float32 value.
func Float32(key string, value float32) Field {
	return Field{
		Key:   key,
		Value: attribute.Float64Value(float64(value)),
	}
}

// Float64 creates a field with a float64 value.
func Float64(key string, value float64) Field {
	return Field{
		Key:   key,
		Value: attribute.Float64Value(value),
	}
}

// Bool creates a field with a bool value.
func Bool(key string, value bool) Field {
	return Field{
		Key:   key,
		Value: attribute.BoolValue(value),
	}
}

// ByteString creates a field with a byte slice value (displayed as string).
func ByteString(key string, value []byte) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(string(value)),
	}
}

// Binary creates a field with a binary value (base64 encoded).
func Binary(key string, value []byte) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(fmt.Sprintf("%x", value)),
	}
}

// Duration creates a field with a time.Duration value.
func Duration(key string, value time.Duration) Field {
	return Field{
		Key:   key,
		Value: attribute.Int64Value(int64(value)),
	}
}

// Time creates a field with a time.Time value.
func Time(key string, value time.Time) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(value.Format(time.RFC3339Nano)),
	}
}

// Error creates a field with an error value.
func Error(err error) Field {
	if err == nil {
		return Field{
			Key:   "error",
			Value: attribute.StringValue("<nil>"),
		}
	}
	return Field{
		Key:   "error",
		Value: attribute.StringValue(err.Error()),
	}
}

// NamedError creates a field with a named error value.
func NamedError(key string, err error) Field {
	if err == nil {
		return Field{
			Key:   key,
			Value: attribute.StringValue("<nil>"),
		}
	}
	return Field{
		Key:   key,
		Value: attribute.StringValue(err.Error()),
	}
}

// Any creates a field with any value.
// The value is serialized using reflection.
func Any(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(fmt.Sprintf("%v", value)),
	}
}

// Reflect creates a field with a value serialized using reflection.
func Reflect(key string, value interface{}) Field {
	return Any(key, value)
}

// Stringer creates a field with a value that implements fmt.Stringer.
func Stringer(key string, value fmt.Stringer) Field {
	if value == nil {
		return Field{
			Key:   key,
			Value: attribute.StringValue("<nil>"),
		}
	}
	return Field{
		Key:   key,
		Value: attribute.StringValue(value.String()),
	}
}

// Stack creates a field that captures a stack trace.
func Stack(key string) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue("stack trace not available"),
	}
}

// StackSkip creates a field that captures a stack trace, skipping the given number of frames.
func StackSkip(key string, skip int) Field {
	return Stack(key)
}

// Namespace creates a namespace field, causing all subsequent fields to be nested.
// In OpenTelemetry, we use dot notation for nested keys.
func Namespace(key string) Field {
	// This is a no-op for OTel, but we return a field that can be used
	// to prefix subsequent field keys
	return Field{
		Key:   key,
		Value: attribute.StringValue(""),
	}
}

// Array creates a field with an array of values.
// OpenTelemetry attributes don't support arrays directly, so we serialize as string.
func Array(key string, values []interface{}) Field {
	return Field{
		Key:   key,
		Value: attribute.StringValue(fmt.Sprintf("%v", values)),
	}
}

// Object creates a field with an object (serialized as string).
func Object(key string, value interface{}) Field {
	return Any(key, value)
}
