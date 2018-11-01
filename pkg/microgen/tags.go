package microgen

import (
	"strings"
)

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
	return strings.Join(x, "\n\t ")
}

func FetchTags(docs []string, prefix string) TagsSet {
	tags := make(TagsSet)
	for _, comment := range docs {
		if !strings.HasPrefix(comment, prefix) {
			continue
		}
		command := strings.SplitN(comment[len(prefix):], ":", 1)
		if len(command[0]) == 0 {
			continue
		}
		tags.Add(command[0], strings.Split(command[1], ",")...)
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
