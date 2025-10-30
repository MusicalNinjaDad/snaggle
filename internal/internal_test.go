package internal

import (
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

	comment, err := GetDocComment(src)
	Assert.NoError(err)
	Assert.Equal(expected, comment)
}
