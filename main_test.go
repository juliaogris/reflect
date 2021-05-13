package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	err := run(&config{})
	require.NoError(t, err)
}
