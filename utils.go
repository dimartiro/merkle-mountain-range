package mmr

import (
	"github.com/tidwall/btree"
)

type MemStorage struct {
	storage *btree.Map[uint64, MMRElement]
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		storage: btree.NewMap[uint64, MMRElement](0),
	}
}

func (s *MemStorage) getElement(pos uint64) (*MMRElement, error) {
	if element, ok := s.storage.Get(pos); ok {
		return &element, nil
	}
	return nil, nil
}

func (s *MemStorage) append(pos uint64, elements []MMRElement) error {
	for i, element := range elements {
		s.storage.Set(pos+uint64(i), element)
	}
	return nil
}

func NewInMemMMR(merger MMRMergeFunction) *MMR {
	return NewMMR(0, NewMMRBatch(NewMemStorage()), merger)
}
