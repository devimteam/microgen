package strings

import (
	"testing"
)

var strs = []string{
	"stringService",
	"StringService",
	"Stringservice",
	"string_service",
	"String_service",
	"string_Service",
	"JSONService",
	"jsonService",
	"JSONServicE",
}

func TestToLower(t *testing.T) {
	anss := []string{
		"stringService",
		"stringService",
		"stringservice",
		"string_service",
		"string_service",
		"string_Service",
		"jsonService",
		"jsonService",
		"jsonServicE",
	}
	if len(strs) != len(anss) {
		t.Fatal("len(strs) != len(anss)")
	}

	for i, s := range strs {
		if ToLower(s) != anss[i] {
			t.Error(i+1, "(", strs[i], "):", ToLower(s), "!=", anss[i])
		}
	}
}

func TestLastUpperOrFirst(t *testing.T) {
	anss := []string{
		"S",
		"S",
		"S",
		"s",
		"S",
		"S",
		"S",
		"S",
		"E",
	}
	if len(strs) != len(anss) {
		t.Fatal("len(strs) != len(anss)")
	}

	for i, s := range strs {
		if LastUpperOrFirst(s) != anss[i] {
			t.Error(i+1, "(", strs[i], "):", LastUpperOrFirst(s), "!=", anss[i])
		}
	}
}
