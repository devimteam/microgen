package svc

import "context"

// This is an interface of the service.
// Yay
type StringService interface {
	Uppercase(ctx context.Context, in, in2 string, in3 int) (c context.Context, err error)
	Lowercase(ctx context.Context, in string) (out string, err error)
}
