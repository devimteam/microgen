package microgen

type Plugin interface {
	Generate(ctx Context) (Context, error)
}
