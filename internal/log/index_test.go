package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndex(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "index_test_foo")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	fmt.Println("Testing with file", f.Name())

	c := Config{}
	c.Segment.MaxIndexBytes = 1024

	idx, err := newIndex(f, c)
	require.NoError(t, err)

	_, _, err = idx.Read(-1)
	// require.NoError(t, err) // FIXME If the file is new, it will have size 0 and thus return io.EOF?
	require.Equal(t, f.Name(), idx.Name())

	entries := []struct {
		Off uint32
		Pos uint64
	}{
		{Off: 0, Pos: 0},
		{Off: 1, Pos: 10},
	}

	for _, want := range entries {
		err = idx.Write(want.Off, want.Pos)
		require.NoError(t, err)

		_, pos, err := idx.Read(int64(want.Off))
		require.NoError(t, err)
		require.Equal(t, want.Pos, pos)
	}

	// Test reading the last entry.
	off, pos, err := idx.Read(-1) // size will be 24, after 2 entries have been appended to the index.
	require.NoError(t, err)
	require.Equal(t, uint32(1), off)
	require.Equal(t, entries[1].Pos, pos)

	// index and scanner should error when reading past exsting entries
	_, _, err = idx.Read(int64(len(entries))) // Read past the post
	require.Equal(t, io.EOF, err)

	// TODO This will return "FlushFileBuffers: The handle is invalid"?
	err = idx.Close()
	require.NoError(t, err)

	// index should build its state from existing file
	f, _ = os.OpenFile(f.Name(), os.O_RDWR, 0644)
	idx, err = newIndex(f, c)
	require.NoError(t, err)

	for _, want := range entries {
		_, pos, err := idx.Read(int64(want.Off))
		require.NoError(t, err)
		require.Equal(t, want.Pos, pos)
	}

	// FIXME These fail, I get back 0, 0 from a -1 Read? But everything works, above.
	off, pos, err = idx.Read(-1) // read the last entry I think?
	require.NoError(t, err)
	require.Equal(t, uint32(1), off)
	require.Equal(t, entries[1].Pos, pos)

}
