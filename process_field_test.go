package secretstruct

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gocloud.dev/runtimevar"
	_ "gocloud.dev/runtimevar/constantvar"
)

func TestProcessField_Error(t *testing.T) {
	ctx := context.Background()
	field := "test"
	info := varInfo{
		Name:  "test",
		URL:   "constant://?val=test-string-value-tag",
		Field: reflect.ValueOf(&field),
	}

	err := processField(ctx, info, func(_ context.Context, _ *runtimevar.Variable) (runtimevar.Snapshot, error) {
		return runtimevar.Snapshot{}, errors.New("test error")
	})

	assert.EqualError(t, err, "test error")
}

func TestProcessField_NilValue(t *testing.T) {
	ctx := context.Background()
	field := "test"
	info := varInfo{
		Name:  "test",
		URL:   "constant://?val=test-string-value-tag",
		Field: reflect.ValueOf(&field),
	}

	err := processField(ctx, info, func(_ context.Context, _ *runtimevar.Variable) (runtimevar.Snapshot, error) {
		return runtimevar.Snapshot{}, nil
	})

	assert.True(t, errors.Is(err, ErrInvalidLatestValue))
}

func TestProcessField_InvalidValueType(t *testing.T) {
	ctx := context.Background()
	field := "test"
	info := varInfo{
		Name:  "test",
		URL:   "constant://?val=test-string-value-tag",
		Field: reflect.ValueOf(&field),
	}

	err := processField(ctx, info, func(_ context.Context, _ *runtimevar.Variable) (runtimevar.Snapshot, error) {
		return runtimevar.Snapshot{
			Value: 123,
		}, nil
	})

	assert.EqualError(t, err, "can't type assert value of type int, must be []byte or string: type mismatch")
	assert.True(t, errors.Is(err, ErrTypeMismatch))
}
