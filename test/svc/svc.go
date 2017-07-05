package svc

// This is an interface of the service.
// Yay
type StringService interface {
	Uppercase(in, in2 string, in3 int) (outWord string, err error)
	Lowercase(in string) (out string, err error)
}
