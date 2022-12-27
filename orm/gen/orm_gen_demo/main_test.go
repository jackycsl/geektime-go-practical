package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_gen(t *testing.T) {
	buffer := &bytes.Buffer{}
	gen(buffer, "testdata/user.go")
	assert.Equal(t, `package testdata

import (
		"github.com/jackycsl/geektime-go-practical/orm/"

		"database/sql"

)
`, buffer.String())
}

func Test_genFile(t *testing.T) {
	f, err := os.Create("testdata/user.gen.go")
	require.NoError(t, err)
	err = gen(f, "testdata/user.go")
	require.NoError(t, err)
}
