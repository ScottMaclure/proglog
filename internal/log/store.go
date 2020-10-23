package log

// TODO Why the mixed use of uint64 and int64?

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian // the encoding to persist record sizes and index entries
)

const (
	lenWidth = 8 // number of bytes to store record's length
)

// Where we're storing Records
type store struct {
	*os.File // wrapper around a file
	mu       sync.Mutex
	buf      *bufio.Writer
	size     uint64
}

func newStore(f *os.File) (*store, error) {
	fp, err := os.Stat(f.Name()) // use Stat in case we're creating from an existing file with data - i.e. restarting a service.
	if err != nil {
		return nil, err
	}

	size := uint64(fp.Size())

	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append adds a Record's data to the Store.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size // Position in store where we'll hold the Record in its File.

	// Write the length of the Record, so when reading back, we know how many bytes to read.
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// Perf gain using buffered writer rather than direct write - helpful for lots of tiny writes.
	numBytesWritten, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	numBytesWritten += lenWidth              // What's this for?
	s.size += uint64(numBytesWritten)        // Update the size of the store.
	return uint64(numBytesWritten), pos, nil // Return number of bytes written (simliar to other Go APIs) & position in store.
}

// Read returns record stored at pos.
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Flush the writer buffer, in case we're trying to read from same data.
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// Find out how many bytes to read to get the whole record.
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// Fetch the record.
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return b, nil
}

// ReadAt reads the length of p bytes starting from an offset.
// Implements io.ReaderAt interface method on the store type.
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close perists buffered data before closing the file.
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}
