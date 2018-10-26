package microgen

type Plugin interface {
	Generate(ctx Context, args []byte) (Context, error)
}
