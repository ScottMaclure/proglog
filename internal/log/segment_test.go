package log

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	api "github.com/ScottMaclure/proglog/api/v1"
	"github.com/stretchr/testify/require"
)

func TestSegment(t *testing.T) {
	dir, _ := ioutil.TempDir("", "segment-test")
	defer os.RemoveAll(dir) // clean up any previous test files, to guarantee cleanness.

	want := &api.Record{Value: []byte("hello world")}

	// Setup the Config used for test purposes - small segment sizes.
	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c) // baseOffset of 16 bytes?
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset, s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(want) // can use string(want.Value) in debugger to get "hello world".
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)
		// FIXME I can compare Values just fine, but I can't compare the protobuf messages - sizeCache is different?
		require.Equal(t, string(want.Value), string(got.Value)) // "hello world" comparison.
		// require.Equal(t, want, got)                             // FIXME this fails on - sizeCache: (int32) 15, + sizeCache: (int32) 0,
	}

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	// maxed index
	require.True(t, s.IsMaxed())

	c.Segment.MaxStoreBytes = uint64(len(want.Value) * 3)
	c.Segment.MaxIndexBytes = 1024

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	// maxed store
	require.True(t, s.IsMaxed())

	err = s.Remove() // FIXME Will get the index's issues - FlushFileBuffers: The handle is invalid, etc.
	require.NoError(t, err)

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())

}
