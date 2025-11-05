package internal

import (
	"bytes"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
		Offset:   28,
		Line:     3,
		Column:   1,
	}

	expectedEnd := token.Position{
		Filename: src,
		Offset:   83,
		Line:     7,
		Column:   3,
	}

	comment, err := getDocComment(src)
	Assert.NoError(err)
	Assert.Equal(expected, comment.Text)
	Assert.Equal(expectedStart, comment.Start)
	Assert.Equal(expectedEnd, comment.End)
}

func TestSetDocComment(t *testing.T) {
	Assert := assert.New(t)
	tmp := t.TempDir()
	original := TestdataPath("hello/hello.go")
	src := filepath.Join(tmp, filepath.Base(original))

	err := Copy(original, src, NewFileLock())
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}

	// ReadAll seems to strip the empty line at EOF ;(
	origCode := `//go:generate echo "hello"

/*
A tiny binary to use for tests

Just says "hello"
*/
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	copied, err := os.Open(src)
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}
	defer func() {
		err = copied.Close()
		if !Assert.ErrorIs(err, fs.ErrClosed) {
			Assert.FailNow("")
		}
	}()

	code, err := io.ReadAll(copied)
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}

	err = copied.Close()
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}

	Assert.Equalf(origCode, string(code), "%s has been changed", original)

	newcomment := "A tiny binary to use for tests\n"
	newcomment += "\n"
	newcomment += "\n"
	newcomment += "Updated comment:\n"
	newcomment += "\n"
	newcomment += "It just says hello\n"

	expected := `//go:generate echo "hello"

/*
A tiny binary to use for tests


Updated comment:

It just says hello
*/
package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

	err = SetDocComment(src, newcomment)
	Assert.NoError(err)

	updated, err := os.Open(src)
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}
	defer func() {
		err = updated.Close()
		if !Assert.NoError(err) {
			Assert.FailNow("")
		}
	}()

	newCode, err := io.ReadAll(updated)
	if !Assert.NoError(err) {
		Assert.FailNow("")
	}

	Assert.NoError(err)
	Assert.Equal(expected, string(newCode))
}

func TestReplaceBetween(t *testing.T) {
	Assert := assert.New(t)

	original := bytes.NewReader([]byte(`Line 1
Line 2

This line ends with something interesting
This stuff
Will all
get changed
Keep me
and me
`))

	replacementText := []byte("new text\nhas been added\n")

	expected := []byte(`Line 1
Line 2

This line ends with something interesting
new text
has been added
Keep me
and me
`)

	updated := ReplaceBetween(original, "something interesting", "Keep me", replacementText)

	Assert.Equal(string(expected), string(updated))
}
