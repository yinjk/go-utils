/**
 * 同步hash map：该数据结构是一个线程安全的HashMap, 底层使用分段锁的机制实现
 * 由于将整个map分割成多段，而每次加锁只用加锁有数据的那一段，从而减小了锁的粒度，也提高了性能。
 * 使用场景：适用于一切线程安全map的场景。
 * @author yinjk
 * @create 2019-06-19 9:43
 */
package maps

import (
	"github.com/yinjk/go-utils/pkg/utils/collection/list"
	"github.com/yinjk/go-utils/pkg/utils/collection/set"
	"github.com/mitchellh/hashstructure"
	"sync"
)

const (
	_defaultSSize = 16
)

type Segment struct {
	sync.Mutex //继承了一把锁，这样Segment就具备了锁的功能
	data       map[interface{}]interface{}
}

type SegmentHashMap struct {
	sync.Mutex
	table []*Segment
}

func NewSegmentHashMap() *SegmentHashMap {
	s := &SegmentHashMap{
		table: make([]*Segment, _defaultSSize, _defaultSSize),
	}
	for i := range s.table {
		s.table[i] = &Segment{
			data: make(map[interface{}]interface{}),
		}
	}
	return s
}

func (m *SegmentHashMap) Put(key interface{}, value interface{}) {
	m.putVal(key, value, false)
}

func (m *SegmentHashMap) PutIfAbsent(key interface{}, value interface{}) bool {
	return m.putVal(key, value, true)
}

func (m *SegmentHashMap) PutAll(maps map[interface{}]interface{}) {
	for k, v := range maps {
		m.Put(k, v)
	}
}

func (m *SegmentHashMap) Get(key interface{}) interface{} {
	keyHash := hash(key)
	segmentIndex := keyHash % uint64(_defaultSSize)
	segment := m.table[segmentIndex]
	return segment.get(key)
}

func (m *SegmentHashMap) GetOrDefault(key interface{}, defaultVal interface{}) (value interface{}) {
	if value = m.Get(key); value == nil {
		return defaultVal
	}
	return value
}

func (m *SegmentHashMap) Remove(key interface{}) (old interface{}) {
	keyHash := hash(key)
	segmentIndex := keyHash % uint64(_defaultSSize)
	segment := m.table[segmentIndex]
	return segment.remove(key)
}

func (m *SegmentHashMap) KeySet(key interface{}) (keys *set.HashSet) {
	panic("implement me")
}

func (m *SegmentHashMap) Values(key interface{}) (values list.List) {
	panic("implement me")
}

//Clear
func (m *SegmentHashMap) Clear() {
	for _, v := range m.table {
		v.clear()
	}
}

//Size 计算map的大小，先采用不加锁的方式，连续计算两次元素个数，如果两次结果相同，表示结果是正确的，如果两次结果不同，对所有segment加锁，重新计算
func (m *SegmentHashMap) Size() int {
	var (
		preSize = 0
		size    = 0
	)
	for _, v := range m.table {
		preSize += v.size()
	}
	for _, v := range m.table {
		size += v.size()
	}
	if preSize == size {
		return size
	}
	for _, s := range m.table {
		s.Lock()
	}
	defer func() {
		for _, s := range m.table {
			s.Unlock()
		}
	}()
	size = 0
	for _, s := range m.table {
		size += s.size()
	}
	return size
}

func (m *SegmentHashMap) putVal(key interface{}, value interface{}, onlyIfAbsent bool) bool {
	keyHash := hash(key)
	segmentIndex := keyHash % uint64(_defaultSSize)
	segment := m.table[segmentIndex]
	segment.Lock()
	defer segment.Unlock()
	return segment.putVal(key, value, onlyIfAbsent)
}

func (s *Segment) size() int {
	return len(s.data)
}

func (s *Segment) putVal(key, value interface{}, onlyIfAbsent bool) bool {
	if _, ok := s.data[key]; ok { //原来有数据
		return false
	}
	s.data[key] = value
	return true
}

func (s *Segment) get(key interface{}) (value interface{}) {
	s.Lock()
	defer s.Unlock()
	return s.data[key]
}

func (s *Segment) remove(key interface{}) (old interface{}) {
	s.Lock()
	defer s.Unlock()
	old = s.data[key]
	delete(s.data, key)
	return old
}

func (s *Segment) clear() {
	s.Lock()
	defer s.Unlock()
	s.data = make(map[interface{}]interface{})
}

func hash(bean interface{}) uint64 {
	if u, e := hashstructure.Hash(bean, nil); e != nil {
		return 0
	} else {
		return u
	}
}
