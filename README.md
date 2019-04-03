# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/) (but not only), by specified service interface.  
The goal is to generate code for service which not fun to write but it should be written.  
Third-party or your own plugins can be used for extending current realization.  

## Install
```
go get -u gopkg.in/devimteam/microgen.v1/cmd/microgen
```

Note: If you have problems with building microgen, please, use go modules or install [dep](https://github.com/golang/dep) and use `dep ensure` command to install correct versions of dependencies ([#29](https://github.com/cv21/microgen/issues/29)).

## Usage
``` sh
microgen [OPTIONS] [interface name]
```

## Example
You may find examples in `examples` directory.

Follow this short guide to try microgen tool.

1. Create file `service.go` inside GOPATH and add code below.
```go
package stringsvc

import (
	"context"

	"github.com/cv21/microgen/example/svc/entity"
)

// @microgen middleware, logging, grpc, http, recovering, main
// @protobuf github.com/cv21/microgen/example/protobuf
type StringService interface {
	// @logs-ignore ans, err
	Uppercase(ctx context.Context, stringsMap map[string]string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
	// @logs-len comments
	TestCase(ctx context.Context, comments []*entity.Comment) (tree map[string]int, err error)
}
```
2. Open command line next to your `service.go`.
3. Enter `microgen`. __*__
4. You should see something like that:
```
@microgen 0.5.0
all files successfully generated
```
5. Now, add and generate protobuf file (if you use grpc transport) and write transport converters (from protobuf/json to golang and _vise versa_).
6. Use endpoints in your `package main` or wherever you want. (tag `main` generates some code for `package main`)

__*__ `GOPATH/bin` should be in your PATH.

## Interface declaration rules
For correct generation, please, follow rules below.

General:
* Interface should be valid golang code.
* All interface method's arguments and results should be named and should be different (name duplicating unacceptable).
* First argument of each method should be of type `context.Context` (from [standard library](https://golang.org/pkg/context/)).
* Last result should be builtin `error` type.
---
GRPC and Protobuf:  
* Name of _protobuf_ service should be the same, as interface name.
* Function names in _protobuf_ should be the same, as in interface.
* Message names in _protobuf_ should be named `<FunctionName>Request` or `<FunctionName>Response` for request/response message respectively.
* Field names in _protobuf_ messages should be the same, as in interface methods (_protobuf_ - snake_case, interface - camelCase).
---
