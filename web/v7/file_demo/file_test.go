package file_demo

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {

	fmt.Println(os.Getwd())

	f, err := os.Open("testdata/my_file.txt")
	require.NoError(t, err)
	data := make([]byte, 64)
	n, err := f.Read(data)
	fmt.Println(n)
	require.NoError(t, err)

	n, err = f.WriteString("hello")
	fmt.Println(n)
	fmt.Println(err)
	require.Error(t, err)
	f.Close()

	f, err = os.OpenFile("testdata/my_file.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	require.NoError(t, err)
	n, err = f.WriteString("hello")
	fmt.Println(n)
	fmt.Println(err)
	f.Close()

	f, err = os.Create("testdata/my_file_copy.txt")
	require.NoError(t, err)
	n, err = f.WriteString("hello, world")
	fmt.Println(n)
	require.NoError(t, err)
}
