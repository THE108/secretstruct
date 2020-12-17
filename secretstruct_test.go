package secretstruct_test

import (
	"context"
	"errors"
	"testing"

	"github.com/THE108/secretstruct"
	"github.com/stretchr/testify/assert"
	_ "gocloud.dev/runtimevar/constantvar"
)

func strPtr(s string) *string {
	return &s
}

func TestProcess(t *testing.T) {
	testStruct := struct {
		FieldSelf      string `secretstruct:"self"`
		FieldTag       string `secretstruct:"constant://?val=test-string-value-tag"`
		FieldIgnored   string `secretstruct:"-"`
		FieldWithNoTag string
		PtrFieldSelf   *string `secretstruct:"self"`
		PtrFieldTag    *string `secretstruct:"constant://?val=test-ptr-string-value-tag"`
		PtrFieldTagNil *string `secretstruct:"constant://?val=test-ptr-string-value-tag-nil"`
		unexported     string  `secretstruct:"constant://?val=test-unexported"`
	}{
		FieldSelf:      "constant://?val=test-string-value-self",
		FieldTag:       "",
		FieldIgnored:   "ignored",
		FieldWithNoTag: "no tag",
		PtrFieldSelf:   strPtr("constant://?val=test-ptr-string-value-self"),
		PtrFieldTag:    strPtr(""),
	}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.NoError(t, err)

	assert.Equal(t, "test-string-value-self", testStruct.FieldSelf)
	assert.Equal(t, "test-string-value-tag", testStruct.FieldTag)
	assert.Equal(t, "ignored", testStruct.FieldIgnored)
	assert.Equal(t, "no tag", testStruct.FieldWithNoTag)
	assert.Equal(t, "test-ptr-string-value-self", *testStruct.PtrFieldSelf)
	assert.Equal(t, "test-ptr-string-value-tag", *testStruct.PtrFieldTag)
	assert.NotNil(t, testStruct.PtrFieldTagNil)
	assert.Equal(t, "test-ptr-string-value-tag-nil", *testStruct.PtrFieldTagNil)
}

func TestProcess_SelfNilPtr(t *testing.T) {
	testStruct := struct {
		PtrFieldSelfNil *string `secretstruct:"self"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.EqualError(t, err, `processing field "PtrFieldSelfNil" failed: open runtimevar.Variable: no scheme in URL ""`)
}

func TestProcess_Embedded(t *testing.T) {
	type EmbeddedStruct struct {
		EmbeddedFieldSelf string `secretstruct:"self"`
		EmbeddedFieldTag  string `secretstruct:"constant://?val=test-embedded-string-value-tag"`
	}

	testStruct := struct {
		FieldSelf string `secretstruct:"self"`
		FieldTag  string `secretstruct:"constant://?val=test-string-value-tag"`
		EmbeddedStruct
	}{
		FieldSelf: "constant://?val=test-string-value-self",
		FieldTag:  "",
		EmbeddedStruct: EmbeddedStruct{
			EmbeddedFieldSelf: "constant://?val=test-embedded-string-value-self",
			EmbeddedFieldTag:  "",
		},
	}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.NoError(t, err)

	assert.Equal(t, "test-string-value-self", testStruct.FieldSelf)
	assert.Equal(t, "test-string-value-tag", testStruct.FieldTag)
	assert.Equal(t, "test-embedded-string-value-self", testStruct.EmbeddedStruct.EmbeddedFieldSelf)
	assert.Equal(t, "test-embedded-string-value-tag", testStruct.EmbeddedStruct.EmbeddedFieldTag)
}

func TestProcess_Inner(t *testing.T) {
	type InnerStruct struct {
		EmbeddedFieldSelf string `secretstruct:"self"`
		EmbeddedFieldTag  string `secretstruct:"constant://?val=test-embedded-string-value-tag"`
	}

	testStruct := struct {
		FieldSelf          string `secretstruct:"self"`
		FieldTag           string `secretstruct:"constant://?val=test-string-value-tag"`
		InnerStruct        InnerStruct
		InnerStructIgnored InnerStruct `secretstruct:"-"`
	}{
		FieldSelf: "constant://?val=test-string-value-self",
		FieldTag:  "",
		InnerStruct: InnerStruct{
			EmbeddedFieldSelf: "constant://?val=test-embedded-string-value-self",
			EmbeddedFieldTag:  "",
		},
		InnerStructIgnored: InnerStruct{
			EmbeddedFieldSelf: "constant://?val=test-embedded-string-value-self",
			EmbeddedFieldTag:  "test",
		},
	}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.NoError(t, err)

	assert.Equal(t, "test-string-value-self", testStruct.FieldSelf)
	assert.Equal(t, "test-string-value-tag", testStruct.FieldTag)
	assert.Equal(t, "test-embedded-string-value-self", testStruct.InnerStruct.EmbeddedFieldSelf)
	assert.Equal(t, "test-embedded-string-value-tag", testStruct.InnerStruct.EmbeddedFieldTag)
	assert.Equal(t, "constant://?val=test-embedded-string-value-self", testStruct.InnerStructIgnored.EmbeddedFieldSelf)
	assert.Equal(t, "test", testStruct.InnerStructIgnored.EmbeddedFieldTag)
}

func TestProcess_UnsupportedFieldType(t *testing.T) {
	testStruct := struct {
		FieldSelf int `secretstruct:"self"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.EqualError(t, err, `field FieldSelf (int): unsupported field type`)
	assert.True(t, errors.Is(err, secretstruct.ErrUnsupportedFieldType))
}

func TestProcess_UnsupportedFieldTypePtr(t *testing.T) {
	testStruct := struct {
		FieldSelfPtr *int `secretstruct:"self"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.EqualError(t, err, `field FieldSelfPtr (*int): unsupported field type`)
	assert.True(t, errors.Is(err, secretstruct.ErrUnsupportedFieldType))
}

func TestProcess_UnsupportedInnerFieldTypePtr(t *testing.T) {
	testStruct := struct {
		InnerStruct struct {
			FieldSelfPtr *int `secretstruct:"self"`
		}
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.EqualError(t, err, `field FieldSelfPtr (*int): unsupported field type`)
	assert.True(t, errors.Is(err, secretstruct.ErrUnsupportedFieldType))
}

func TestProcess_UnsupportedDecoder(t *testing.T) {
	testStruct := struct {
		FieldSelf string `secretstruct:"constant://?val=12345&decoder=int"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.EqualError(t, err, `processing field "FieldSelf" failed: open variable constant:?val=12345&decoder=int: invalid decoder: unsupported decoder "int"`)
}

func TestProcess_BytesDecoder(t *testing.T) {
	testStruct := struct {
		Field string `secretstruct:"constant://?val=test-bytes&decoder=bytes"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.NoError(t, err)
	assert.Equal(t, "test-bytes", testStruct.Field)
}

func TestProcess_StringDecoder(t *testing.T) {
	testStruct := struct {
		Field string `secretstruct:"constant://?val=test-bytes&decoder=string"`
	}{}

	err := secretstruct.Process(context.Background(), &testStruct)
	assert.NoError(t, err)
	assert.Equal(t, "test-bytes", testStruct.Field)
}

func TestProcess_InvalidStruct(t *testing.T) {
	assert.True(t, errors.Is(secretstruct.Process(context.Background(), nil), secretstruct.ErrInvalidInputParamType))
	assert.True(t, errors.Is(secretstruct.Process(context.Background(), ""), secretstruct.ErrInvalidInputParamType))

	s := "test"
	assert.True(t, errors.Is(secretstruct.Process(context.Background(), &s), secretstruct.ErrInvalidInputParamType))
}
