package svc

// This is an interface of the service.
// Yay
type StringService interface {
	Uppercase(in, in2 string, in3 int) (out string, err error)
}