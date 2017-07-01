package svc

type StringService interface {
	Uppercase(in, in2 string, in3 int) (out string, err error)
}