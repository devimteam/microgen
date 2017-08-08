# Microgen

Tool to generate microservices, based on [go-kit](https://gokit.io/), by specified service interface.

## Usage
TODO
## CLI params
TODO
### Misc

* All interface arguments and results should be named.
* First argument should be `ctx context.Context`.
* Argument's and result's names should be different or have same type.

#### Not allowed names:
```
ctx
req
resp
```