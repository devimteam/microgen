package template

import "context"

const (
	spi                = "SourcePackageImport"
	ael                = "AllowEllipsis"
	mainTagsContextKey = "MainTags"
)

func WithSourcePackageImport(parent context.Context, val string) context.Context {
	return context.WithValue(parent, spi, val)
}

func SourcePackageImport(ctx context.Context) string {
	return ctx.Value(spi).(string)
}

func WithTags(parent context.Context, tt TagsSet) context.Context {
	return context.WithValue(parent, mainTagsContextKey, tt)
}

func Tags(ctx context.Context) TagsSet {
	return ctx.Value(mainTagsContextKey).(TagsSet)
}

type TagsSet map[string]struct{}

func (s TagsSet) Has(item string) bool {
	_, ok := s[item]
	return ok
}

func (s TagsSet) HasAny(items ...string) bool {
	if len(items) == 0 {
		return false
	}
	return s.Has(items[0]) || s.HasAny(items[1:]...)
}

func (s TagsSet) Add(item string) {
	s[item] = struct{}{}
}

func AllowEllipsis(ctx context.Context) bool {
	v, ok := ctx.Value(ael).(bool)
	return ok && v
}
