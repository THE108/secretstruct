[![Build Status](https://github.com/THE108/secretstruct/workflows/Test/badge.svg)](https://github.com/THE108/secretstruct/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/THE108/secretstruct.svg)](https://pkg.go.dev/github.com/THE108/secretstruct)
[![Coverage](https://codecov.io/gh/THE108/secretstruct/branch/main/graph/badge.svg)](https://codecov.io/gh/THE108/secretstruct)

# secretstruct

Golang package to parse secret values from secret providers to struct fields.

This package uses `runtimevar` package from [go-cloud](https://github.com/google/go-cloud) library 
to get secret values from supported stores (see [docs](https://gocloud.dev/howto/runtimevar/)).

## Usage

Annotate a field with `secretstruct` tag to fetch a variable from supported secret providers.

`secretstruct` tag can contain either a URL of a format used by `gocloud.dev/runtimevar` package
(see [runtimevar docs](https://gocloud.dev/howto/runtimevar/)) or `self` which denotes that this field value
contains the URL.

```go
package main

import (
    "context"
    "fmt"

    "github.com/THE108/secretstruct"
    
    // Use blank imports to init providers supported by `github.com/google/go-cloud/runtimevar`.
    _ "gocloud.dev/runtimevar/awsparamstore"
    _ "gocloud.dev/runtimevar/awssecretsmanager"
    _ "gocloud.dev/runtimevar/constantvar"
    _ "gocloud.dev/runtimevar/gcpsecretmanager"
)

type TestStruct struct {
    // This field will be fetched from AWS Secrets Manager (see https://aws.amazon.com/en/secrets-manager/).
    FieldAWSSecretsManager string `secretstruct:"awssecretsmanager://?val=test-string-value-from-aws-secrets-manager"`
    
    // This field will be fetched from GCP Secret Manager (see https://cloud.google.com/secret-manager).
    FieldGCPSecretManager string `secretstruct:"gcpsecretmanager://?val=test-string-value-from-gcp-secret-manager"`
    
    // This field will be fetched using the URL from the current FieldAWSParamStore field value
    // (see struct init below).
    FieldAWSParamStore string `secretstruct:"self"`
}

func main() {
    ctx := context.Background()
    testStruct := TestStruct{
        // This field will be fetched from AWS Param Store.
        FieldAWSParamStore: "awsparamstore://?val=test-string-value-from-aws-param-store",
    }

    // Call Process to fetch all string values marked with `secretstruct` tag.
    if err := secretstruct.Process(ctx, &testStruct); err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("testStruct: %+v\n", testStruct)
}
```

Embedded and internal structs are also supported:

```go
type EmbeddedStruct struct {
    EmbeddedField string `secretstruct:"awssecretsmanager://?val=test-string-value-from-aws-secrets-manager"`
}

type TestStruct struct {
    EmbeddedStruct
    InnerStruct struct {
        FieldAWSSecretsManager string `secretstruct:"awssecretsmanager://?val=test-string-value-from-aws-secrets-manager"`
    }
    FieldGCPSecretManager string `secretstruct:"gcpsecretmanager://?val=test-string-value-from-gcp-secret-manager"`
}
```

To ignore a field use `-` tag value:

```go
type TestStruct struct {
    IgnoredField string `secretstruct:"-"`
}
```

## License

MIT
