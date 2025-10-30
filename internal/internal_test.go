package internal

import (
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFile(t *testing.T) {
	same, err := sameHash(P_hello_pie, P_hello_static)
	assert.NoError(t, err)
	assert.False(t, same)
}

func TestGetDoccomment(t *testing.T) {
	Assert := assert.New(t)

	src := TestdataPath("hello/hello.go")
	expected := "/*\n"
	expected += "A tiny binary to use for tests\n"
	expected += "\n"
	expected += "Just says \"hello\"\n"
	expected += "*/"

	expectedStart := token.Position{
		Filename: src,
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	comment, err := GetDocComment(src)
	Assert.NoError(err)
	Assert.Equal(expected, comment.Text)
	Assert.Equal(expectedStart, comment.Start)
}
