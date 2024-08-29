package mmr

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func mergeNumberHash(left, right MMRElement) (MMRElement, error) {
	hasher, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}

	hasher.Write(left)
	hasher.Write(right)

	return hasher.Sum(nil), nil
}

func hashNumber(number int) MMRElement {
	hasher, _ := blake2b.New256(nil)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(number))
	hasher.Write(buf)
	return hasher.Sum(nil)
}

func TestPush1Elements(t *testing.T) {
	inMemMMR := NewInMemMMR(mergeNumberHash)

	leaf := hashNumber(42)
	_, err := inMemMMR.Push(leaf)
	assert.NoError(t, err)

	root, err := inMemMMR.Root()
	assert.NoError(t, err)

	assert.Equal(t, root, leaf)
}

func TestPush4Elements(t *testing.T) {
	inMemMMR := NewInMemMMR(mergeNumberHash)

	for i := 0; i < 2; i++ {
		leaf := hashNumber(i)
		_, err := inMemMMR.Push(leaf)
		assert.NoError(t, err)
	}

	root, err := inMemMMR.Root()
	assert.NoError(t, err)

	assert.Equal(t, []byte(root), []byte{0x2a, 0x44, 0xf7, 0xc, 0xa4, 0x6b, 0xee, 0x95, 0xa, 0x4b, 0xd3, 0x52, 0x8a, 0x3a, 0x3a, 0x10, 0xc4, 0x3d, 0x19, 0x51, 0x9c, 0xfe, 0x67, 0xc7, 0x93, 0x94, 0x3a, 0x12, 0xfc, 0x7, 0xf7, 0xe7})
}
