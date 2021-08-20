package service

import (
	"gotest.tools/assert"
	"testing"
)

func TestIntersect(t *testing.T) {
	a := []string{"shortenUrl:iiiGGUFrOo",
		"shortenUrl:7XxYzjImrg6",
		"shortenUrl:4oEQByEsvg4",
		"shortenUrl:6dykW3VqLoC",
		"shortenUrl:7In6TtKH1kU",
		"shortenUrl:4fK1MXvjQGi",
		"shortenUrl:1idrdKDoC4U",
	}
	b := []string{"shortenUrl:4oEQByEsvg4"}
	c := intersect(a, b)
	assert.DeepEqual(t, b, c)
}
