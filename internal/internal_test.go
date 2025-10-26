package internal

import (
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFile(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })
	i, _ := strconv.Atoi(log.Prefix())
	log.SetPrefix(strconv.Itoa(i+1) + " ")

	log.Default().Printf("foo")
	same, err := sameHash(P_hello_pie, P_hello_static)
	assert.NoError(t, err)
	assert.False(t, same)
	assert.Equal(t, "foo", stdout.String())
}
