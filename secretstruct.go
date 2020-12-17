package secretstruct

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"go.uber.org/multierr"
	"gocloud.dev/runtimevar"
	"golang.org/x/sync/errgroup"
)

const (
	tagName = "secretstruct"

	selfTagValue = "self"
	stopTagValue = "-"
)

var (
	// ErrInvalidInputParamType indicates that an input param is of the wrong type.
	ErrInvalidInputParamType = errors.New("input param must be a struct pointer")

	// ErrUnsupportedFieldType indicates that a field annotated with `secretstruct` tag is of the wrong type.
	// Could be a string or *string.
	ErrUnsupportedFieldType = errors.New("unsupported field type")

	// ErrInvalidLatestValue indicates that a call to `gocloud.dev/runtimevar#Variable.Latest`
	// method returned nil.
	ErrInvalidLatestValue = errors.New("latest value is nil")

	// ErrTypeMismatch indicates that a call to `gocloud.dev/runtimevar#Variable.Latest`
	// method returned something different from []byte or string.
	ErrTypeMismatch = errors.New("type mismatch")
)

// Process populates the specified struct based on struct field tags.
func Process(ctx context.Context, in interface{}) error {
	infos, err := gatherInfo(in)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)

	for _, info := range infos {
		info := info

		g.Go(func() error {
			if err := processField(ctx, info, getLatest); err != nil {
				return fmt.Errorf("processing field %q failed: %w", info.Name, err)
			}

			return nil
		})
	}

	return g.Wait()
}

// varInfo maintains information about the configuration variable.
type varInfo struct {
	Name  string
	URL   string
	Field reflect.Value
}

// GatherInfo gathers information about the specified struct.
func gatherInfo(spec interface{}) ([]varInfo, error) {
	s := reflect.ValueOf(spec)
	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidInputParamType
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidInputParamType
	}

	typeOfSpec := s.Type()

	var infos []varInfo

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		if !f.CanSet() {
			continue
		}

		structField := typeOfSpec.Field(i)

		val, hasTag := structField.Tag.Lookup(tagName)

		// if this field is explicitly ignored
		if val == stopTagValue {
			continue
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				elemKind := f.Type().Elem().Kind()
				if elemKind != reflect.Struct && elemKind != reflect.String {
					// nil pointer to a non-struct: leave it alone
					break
				}

				// nil pointer to struct or string: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		// recursively process inner struct
		if f.Kind() == reflect.Struct {
			embeddedInfos, err := gatherInfo(f.Addr().Interface())
			if err != nil {
				return nil, err
			}

			infos = append(infos, embeddedInfos...)

			continue
		}

		// if this field does not have the tag
		if !hasTag {
			continue
		}

		// only strings are supported at the moment. []byte and JSON will be supported later
		if f.Kind() != reflect.String {
			return nil, fmt.Errorf("field %s (%s): %w", structField.Name, f.Type(), ErrUnsupportedFieldType)
		}

		// replace "self" special tag value with field current value
		if val == selfTagValue {
			val = f.String()
		}

		infos = append(infos, varInfo{
			Name:  structField.Name,
			URL:   val,
			Field: f,
		})
	}

	return infos, nil
}

type getLatestFunc func(ctx context.Context, v *runtimevar.Variable) (runtimevar.Snapshot, error)

func processField(ctx context.Context, info varInfo, f getLatestFunc) (err error) {
	v, err := runtimevar.OpenVariable(ctx, info.URL)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Append(err, v.Close())
	}()

	latest, err := f(ctx, v)
	if err != nil {
		return err
	}

	return setFieldValue(info.Field, latest.Value)
}

func getLatest(ctx context.Context, v *runtimevar.Variable) (runtimevar.Snapshot, error) {
	return v.Latest(ctx)
}

func setFieldValue(dst reflect.Value, src interface{}) error {
	if src == nil {
		return ErrInvalidLatestValue
	}

	// Default variable decoder is ByteDecoder.
	// Try to type assert to both []byte and string.
	// This is needed to let users to avoid specifying decoder=string param in URL
	// if dst is string and src either string or []byte.
	switch data := src.(type) {
	case []byte:
		dst.SetString(string(data))
	case string:
		dst.SetString(data)
	default:
		return fmt.Errorf("can't type assert value of type %T, must be []byte or string: %w",
			src, ErrTypeMismatch)
	}

	return nil
}
