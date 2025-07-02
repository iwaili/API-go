// bloomfilter.go

package bloomfilter

import (
	"crypto/sha256"
	"encoding/gob"
	"math/big"
	"os"
)

// BloomFilter represents the Bloom filter structure.
type BloomFilter struct {
	Bits *big.Int
	Size uint
	K    uint
}

// NewBloom creates and initializes a new Bloom filter with a specified size and number of hash functions.
func NewBloom(size, k uint) *BloomFilter {
	return &BloomFilter{
		Bits: big.NewInt(0),
		Size: size,
		K:    k,
	}
}

// Add adds an item to the Bloom filter.
func (bf *BloomFilter) Add(item string) {
	hash := sha256.Sum256([]byte(item))
	for i := uint(0); i < bf.K; i++ {
		idx := bf.hashAt(hash, i)
		bf.Bits.SetBit(bf.Bits, int(idx), 1)
	}
}

// Contains checks if an item is present in the Bloom filter.
func (bf *BloomFilter) Contains(item string) bool {
	hash := sha256.Sum256([]byte(item))
	for i := uint(0); i < bf.K; i++ {
		idx := bf.hashAt(hash, i)
		if bf.Bits.Bit(int(idx)) == 0 {
			return false
		}
	}
	return true
}

// Save saves the Bloom filter to a file with a specific extension.
func (bf *BloomFilter) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(bf)
}

// LoadBloom loads a Bloom filter from a file.
func LoadBloom(path string) (*BloomFilter, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var bf BloomFilter
	err = gob.NewDecoder(f).Decode(&bf)
	return &bf, err
}

// hashAt generates a hash value for the given index and returns the position in the Bloom filter.
func (bf *BloomFilter) hashAt(hash [32]byte, i uint) uint {
	start := (i * 4) % 28
	val := (uint(hash[start])<<24 | uint(hash[start+1])<<16 | uint(hash[start+2])<<8 | uint(hash[start+3]))
	return val % bf.Size
}
