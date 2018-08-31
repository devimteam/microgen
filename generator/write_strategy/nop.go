package write_strategy

type nopStrategy struct{}

// Do nothing strategy
func NewNopStrategy() Strategy                    { return nopStrategy{} }
func (s nopStrategy) Write(Renderer) error        { return nil }
func (s nopStrategy) Save(Renderer, string) error { return nil }
