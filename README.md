# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/), by specified service interface.

## Usage
``` sh
microgen [OPTIONS]
```
### Options

| Name        | Default | Description                                                                   |
|:------------|:--------|:------------------------------------------------------------------------------|
| -file*      |         | Relative path to source file with service interface                           |
| -interface* |         | Service interface name in source file                                         |
| -out*       |         | Relative or absolute path to directory, where you want to see generated files |
| -package*   |         | Package path of your service interface source file                            |
| -debug      | false   | Display some debug information                                                |

\* __Required option__

### Misc

* All interface arguments and results should be named.
* First argument should be `ctx context.Context`.
* Argument's and result's names should be different or have same type.

#### Not allowed names:
```
req
resp
```