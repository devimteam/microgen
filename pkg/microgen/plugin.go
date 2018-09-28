package microgen

type Plugin interface {
	Generate(ctx Context, args ...string) (Context, error)
}
