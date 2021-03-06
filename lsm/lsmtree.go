package lsm

import (
	"github.com/Pigeon377/Dandelion/filter"
	"github.com/Pigeon377/Dandelion/sstable"
)

type LSM struct {
	table *sstable.SSTable
	bloom *filter.BloomFilter
}

func NewLSM() *LSM {
	var f *filter.BloomFilter
	if filter.IsBloomFilterPersistenceExist() {
		f = filter.GeneratorFilterFromFile()
	} else {
		f = filter.NewBloomFilter()
	}
	return &LSM{
		table: sstable.NewSSTable(),
		bloom: f,
	}
}

func (lsm *LSM) Get(key int) ([]byte, bool) {
	if lsm.bloom.Get(key) {
		return lsm.table.Get(key)
	} else {
		return nil, false
	}
}

func (lsm *LSM) Put(key int, value []byte) error {
	lsm.bloom.Put(key)
	return lsm.table.Put(key, value)
}

func (lsm *LSM) Update(key int, value []byte) error {
	lsm.bloom.Put(key)
	return lsm.table.Update(key, value)
}

func (lsm *LSM) Delete(key int) error {
	return lsm.table.Delete(key)
}

// Flush
// persistence data to disk
func (lsm *LSM) Flush() error {
	return lsm.table.Flush()
}

// ClearMemory
// flush data to disk
// and clear skiplist in memory
func (lsm *LSM) ClearMemory() error {
	return lsm.table.ClearMemory()
}

func (lsm *LSM) ClearFilter() error {
	return filter.ClearBloomFilter(lsm.bloom)
}
