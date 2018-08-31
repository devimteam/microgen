package internal

import (
	"context"
	"strings"
)

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

type TagsSet map[string][]string

func (s TagsSet) Get(item string) []string {
	return s[item]
}

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

func (s TagsSet) Add(item string, content ...string) {
	s[item] = content
}

func (s TagsSet) String() string {
	x := make([]string, len(s))
	i := 0
	for k, v := range s {
		x[i] = k + ": " + strings.Join(v, " ")
		i++
	}
	return strings.Join(x, "\n")
}

func AllowEllipsis(ctx context.Context) bool {
	v, ok := ctx.Value(ael).(bool)
	return ok && v
}

func FetchTags(docs []string, prefix string) TagsSet {
	tags := make(TagsSet)
	for _, comment := range docs {
		if !strings.HasPrefix(comment, prefix) {
			continue
		}
		command := strings.Split(comment[len(prefix):], " ")
		if len(command[0]) == 0 {
			continue
		}
		tags.Add(command[0], command[1:]...)
	}
	return tags
}

func FetchList(docs []string, prefix string) []string {
	var list []string
	for _, comment := range docs {
		if !strings.HasPrefix(comment, prefix) {
			continue
		}
		list = append(list, strings.Split(comment[len(prefix):], " ")...)
	}
	return list
}
