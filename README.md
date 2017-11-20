# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/), by specified service interface.  
The goal is to generate code for service which not fun to write but it should be written.

## Install
```
go get -u github.com/devimteam/microgen/cmd/microgen
```

## Usage
``` sh
microgen [OPTIONS]
```
microgen tool search in file first `type * interface` with docs, that contains `// @microgen`.

generation parameters provides through ["tags"](#tags) in interface docs after general `// @microgen` tag (space before @ __required__).

### Options

| Name   | Default    | Description                                                                   |
|:------ |:-----------|:------------------------------------------------------------------------------|
| -file  | service.go | Relative path to source file with service interface                           |
| -out   | .          | Relative or absolute path to directory, where you want to see generated files |
| -force | false      | With flag generate stub methods.                                              |
| -help  | false      | Print usage information                                                       |

\* __Required option__

### Markers
Markers is a general tags, that participate in generation process.
Typical syntax is: `// @<tag-name>:`

#### @microgen
Main tag for microgen tool. Microgen scan file for the first interface which docs contains this tag.  
To add templates for generation, add their tags, separated by comma after `@microgen:`
Example:
```go
// @microgen middleware, logging
type StringService interface {
    ServiceMethod()
}
```
#### @protobuf
Protobuf tag is used for package declaration of compiled with `protoc` grpc package.  
Example:
```go
// @microgen grpc-server
// @protobuf github.com/user/repo/path/to/protobuf
type StringService interface {
    ServiceMethod()
}
```
`@protobuf` tag is optional, but required for `grpc`, `grpc-server`, `grpc-client` generation.  
#### @grpc-addr
gRPC address tag is used for gRPC go-kit-based client generation.
Example:
```go
// @microgen grpc
// @protobuf github.com/user/repo/path/to/protobuf
// @grpc-addr some.service.address
type StringService interface {
    ServiceMethod()
}
```
`@grpc-addr` tag is optional, but required for `grpc-client` generation.
#### @force
Use force command for provided tags anyway. Useful, when you want to generate again after changing service methods.
Used by default for `middleware` and `logging`.

### Method's tags
#### @logs-ignore
This tag is used for logging middleware, when some arguments or results should not be logged, e.g. passwords or files.  
If `context.Context` is first argument, it ignored by default.
Provide parameters names, separated by comma, to exclude them from logs.  
Example:
```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

#### @logs-len
This tag is used for logging middleware. It prints length of parameters.
Example:  
```go
// @microgen logging
type FileService interface {
    // @logs-ignore data
    // @logs-len data
    UploadFile(ctx context.Context, name string, data []byte) (link string, err error)
}
```

### Tags
All allowed tags for customize generation provided here.

| Tag         | Description                                                                                                                   |
|:------------|:------------------------------------------------------------------------------------------------------------------------------|
| middleware  | General application middleware interface. Generates every time.                                                               |
| logging     | Middleware that writes to logger all request/response information with handled time. Generates every time.                    |
| recover     | Middleware that recovers panics and writes errors to logger. Generates every time.                                            |
| grpc-client | Generates client for grpc transport with request/response encoders/decoders. Do not generates again if file exist.            |
| grpc-server | Generates server for grpc transport with request/response encoders/decoders. Do not generates again if file exist.            |
| grpc        | Generates client and server for grpc transport with request/response encoders/decoders. Do not generates again if file exist. |
| http-client | Generates client for http transport with request/response encoders/decoders. Do not generates again if file exist.            |
| http-server | Generates server for http transport with request/response encoders/decoders. Do not generates again if file exist.            |
| http        | Generates client and server for http transport with request/response encoders/decoders. Do not generates again if file exist. |

## Example
Follow this short guide to try microgen tool.

1. Create file `service.go` inside GOPATH and add code below.
```go
package stringsvc

import (
    "context"

    drive "google.golang.org/api/drive/v3"
)

// @microgen grpc, middleware, logging
// @protobuf github.com/devimteam/proto-utils
// @grpc-addr test.address
type StringService interface {
    Uppercase(ctx context.Context, str string) (ans string, err error)
    Count(ctx context.Context, text string, symbol string) (count int, positions []int)
}
```
2. Open command line next to your `service.go`.
3. Enter `microgen`. __*__
4. You should see something like that:
```
exchanges.go
endpoints.go
client.go
middleware/middleware.go
middleware/logging.go
transport/grpc/server.go
transport/grpc/client.go
transport/converter/protobuf/endpoint_converters.go
transport/converter/protobuf/type_converters.go
All files successfully generated
```
5. Now, add and generate protobuf file, write converters from protobuf to golang and _vise versa_.
6. Use endpoints and converters in your `package main` or wherever you want.

__*__ `GOPATH/bin` should be in your PATH.

## Interface declaration rules
For correct generation, please, follow rules below.

* All interface method's arguments and results should be named and should be different (name duplicating unacceptable).
* First argument of each method should be of type `context.Context` (from [standard library](https://golang.org/pkg/context/)).
* Last result should should be `error` type.
---
* Name of _protobuf_ service should be the same, as interface name.
* Function names in _protobuf_ should be the same, as in interface.
* Message names in _protobuf_ should be named `<FunctionName>Request` or `<FunctionName>Response` for request/response message respectively.
* Field names in _protobuf_ messages should be the same, as in interface methods (_protobuf_ - snake_case, interface - camelCase).

## Dependency
After generation your service may depend on this packages:
```
    "net/http"      // for http purposes
    "bytes"
    "encoding/json" // for http purposes
    "io/ioutil"
    "strings"
    "net/url"       // for http purposes
    "fmt"
    "context"
    "time"          // for logging

    "google.golang.org/grpc"                    // for grpc purposes
    "golang.org/x/net/context"
    "github.com/go-kit/kit"                     // for grpc purposes
    "github.com/golang/protobuf/ptypes/empty"   // for grpc purposes
```