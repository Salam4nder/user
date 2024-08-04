package server

import (
	"errors"
	"fmt"

	"github.com/Salam4nder/user/proto/gen"
	"go.opentelemetry.io/otel/attribute"
)

type GenReq interface {
	*gen.Credentials | *gen.Number
}

// GenerateSpanAttributes returns span attributes for generated request structs.
// Experimental solution, not the prettiest.
func GenSpanAttributes[T GenReq](param T) ([]attribute.KeyValue, error) {
	if param == nil {
		return nil, errors.New("param is nil")
	}

	switch t := any(param).(type) {
	case *gen.Credentials:
		return []attribute.KeyValue{
			attribute.String("email", t.GetEmail()),
			attribute.Int("password length", len(t.GetPassword())),
		}, nil
	case *gen.Number:
		return []attribute.KeyValue{
			attribute.Int64("number", int64(t.GetNumbers())),
		}, nil
	default:
		return nil, fmt.Errorf("server: span attributes, unsupported type %T", t)
	}
}
