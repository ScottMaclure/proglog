// Segment wraps the Index and Store types, and coordinates operations across the two.
// When the Log appends to the Segment, the Segment adds to the Store and adds an entry to the Index.
// For reads, the Segment looks up the Index and then fetches the data from the Store.

package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/ScottMaclure/proglog/api/v1"
	proto "github.com/golang/protobuf/proto"
)

type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64 // Offsets are for new records.
	config                 Config // For tracking when Segments are filled.
}

// Log calls newSegment when adding a new segment.
func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	var err error

	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		// Create means it will create the file if it doesn't exist.
		// Append means the os will append when writing to file.
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}

	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}

	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}

	// Set offsets ready for next Record.
	if off, _, err := s.index.Read(-1); err != nil {
		s.nextOffset = baseOffset
	} else {
		//  If the index has at least one entry, then that means the offset of the next record written should take
		// the offset at end of the segment, which we get by adding 1 to the base offset and relative offset.
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

// Append writes the Record to the Segment and returns the new offset.
// Two-step process: appends data to store, adds index entry.
// Index offsets are relative to base offset, so subtract the segments next offset from its base offset (both absolute).
// Then increment the next offset in prep for future Append call.
func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}
	if err = s.index.Write(
		// index offsets are relative to base offset
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

// Returns record for given offset.
// Translate absolute index to relative offset to get index entry.
// Then go straight to Record's position in Store, and read the proper data.
func (s *segment) Read(off uint64) (*api.Record, error) {
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

// IsMaxed checks for segment full.
// long logs = segment bytes limit, small logs = index bytes limit.
func (s *segment) IsMaxed() bool {
	return s.store.size >= s.config.Segment.MaxStoreBytes ||
		s.index.size >= s.config.Segment.MaxIndexBytes
}

// Close the Index and Store.
func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

// Remove closes the Segment and removes the Index and Store.
func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

// nearestMultiple used stay under disk capacity.
func nearestMultiple(j, k uint64) uint64 {
	if j >= 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k

}
