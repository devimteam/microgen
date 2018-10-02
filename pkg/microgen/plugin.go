package microgen

import "encoding/json"

type Plugin interface {
	Generate(ctx Context, args json.RawMessage) (Context, error)
}
