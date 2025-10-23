package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXxx(t *testing.T) {
	cmd := exec.Command("./snaggle")
	err := cmd.Run()
	assert.NoError(t, err)
}
