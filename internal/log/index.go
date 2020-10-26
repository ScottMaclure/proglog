// Index for the Store.
// On start, service needs to know the offset to set on next Record appended to Log.
// It looks at last entry of the Index - reading the last 12 bytes in the file.
// This is complicated by growing the file and mmap'ing it (can't resize after mmaping them, now or never).
// Grow by appending empty space, which prevents the service from restarting properly - that's why we shrink back down on Close.
// SM Q: What happens if the service crashes or is killed before it can do this? You're screwed?
// A: Perform "sanity check" when service (re)starts to find corrupted data.
// If you find corruption, you rebuild the data (if you can), or replicate it from an uncorrupted source.
// The below code does NOT handle ungraceful shutdowns, for brevity.

package log

import (
	"io"
	"os"

	"github.com/tysontate/gommap"
)

// SM: I used const instead of var, because the book says they're constants.
const (
	// The *Width constants define the number of bytes that make up each index entry.
	// index entries contain two fields: the record's offset and its position in the store file.
	// offsets stored as uint32s and positions as uint64s, so they take up 4 and 8 bytes respectively.
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth // Used to jump straight to position of entry given its offset, since pos = off+pos widths.
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
}

// Create index and save current size of file, so we can track data as we add more entries.
// Grow file to max index size before memory-mapping the file.
func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	idx.size = uint64(fi.Size())

	if err = os.Truncate(
		f.Name(), int64(c.Segment.MaxIndexBytes),
	); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),
		gommap.PROT_READ|gommap.PROT_WRITE,
		gommap.MAP_SHARED,
	); err != nil {
		return nil, err
	}

	return idx, nil
}

// Read takes "in" an offset and returns the associated Record's position in the Store.
// The given offset is relative to the Segment's base offset - 0 is always the offset of the index's first entry, 1 is the second, etc.
// Relative offsets to reduce size of indexes by storing offsets as uint32s. Absolute offsets would require uint64s. Think trillions of records a day for scale. Even billions a day isn't unfeasible.
func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		out = uint32((i.size / entWidth) - 1) // last entry?
	} else {
		out = uint32(in)
	}

	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}

	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])

	return out, pos, nil
}

// Write appends given offset and position to the index.
func (i *index) Write(off uint32, pos uint64) error {
	// Validate we have space.
	if uint64(len(i.mmap)) < i.size+entWidth {
		return io.EOF // Need a new Segment?
	}

	// Encode offset & position and write them to mmap file.
	enc.PutUint32(i.mmap[i.size:i.size+offWidth], off)
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)

	// Increment for next write.
	i.size += uint64(entWidth)

	return nil
}

// Name returns the file path of the index.
func (i *index) Name() string {
	return i.file.Name()
}

// Close makes sure the mmap'ed file has persisted data to file before closing. For restarts, etc.
func (i *index) Close() error {
	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		// FIXME During unit testing, this produces an error (Win10).
		// "FlushFileBuffers: The handle is invalid"
		return err
	}

	if err := i.file.Sync(); err != nil {
		return err
	}

	// Shrink back down, to enable re-start to work again (i.e. Graceful shutdown).
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}
