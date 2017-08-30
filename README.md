# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/), by specified service interface.

## Install
```
go get -u github.com/devimteam/microgen/cmd/microgen
```

## Usage
``` sh
microgen [OPTIONS]
```
microgen is stable, so you can generate without flag `-init` any time your interface changed (e.g. added new method)
### Options

| Name        | Default          | Description                                                                   |
|:------------|:-----------------|:------------------------------------------------------------------------------|
| -file*      |                  | Relative path to source file with service interface                           |
| -interface* |                  | Service interface name in source file                                         |
| -out        | writes to stdout | Relative or absolute path to directory, where you want to see generated files |
| -package*   |                  | Package path of your service interface source file                            |
| -debug      | false            | Display some debug information                                                |
| -grpc       | false            | Render client, server and converters for gRPC protocol                        |
| -init       | false            | With flag `-grpc` generate stub methods for converters                        |

\* __Required option__

## Example
Follow this short guide to try microgen tool.

1. Create file `service.go` inside GOPATH and add code below.
``` golang
package stringsvc

import (
	"context"

	drive "google.golang.org/api/drive/v3"
)

type StringService interface {
	Uppercase(ctx context.Context, str string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int)
	TestCase(ctx context.Context, comments []*drive.Comment) (err error)
}
```
2. Open command line next to your `service.go`.
3. Enter `microgen -file ./service.go -interface StringService -out . -grpc -init`. __*__
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

### Interface declaration rules
For correct generation, please, follow rules below.

* All interface method's arguments and results should be named and should be different (name duplicating unacceptable).
* First argument of each method should be of type `context.Context` (from [standard library](https://golang.org/pkg/context/)).
* Method results should contain at least one variable of `error` type.
* [Some names](#not-allowed-names) are not allowed to be an argument or result.
* Field names in _protobuf_ messages should be the same, as in interface methods (_protobuf_ - snake_case, interface - camelCase).

#### Not allowed names:
```
req
request
resp
response
```

### Misc

Microgen uses __0.9.*__ version of [devimteam/go-kit](https://github.com/devimteam/go-kit)