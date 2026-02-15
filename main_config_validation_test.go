package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestValidateConfigReturnsErrorWhenTokenIsEmpty(t *testing.T) {
	t.Parallel()

	err := validateConfig("", "https://api.bexio.com")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BEXIO_API_TOKEN")
}
