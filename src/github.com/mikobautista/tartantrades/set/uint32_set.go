package set

type IntSet struct {
	set map[uint32]bool
}

func NewUint32Set() *IntSet {
	return &IntSet{make(map[uint32]bool)}
}

func (set *IntSet) Add(i uint32) bool {
	_, found := set.set[i]
	set.set[i] = true
	return !found //False if it existed already
}

func (set *IntSet) Get(i uint32) bool {
	_, found := set.set[i]
	return found //true if it existed already
}

func (set *IntSet) Remove(i uint32) {
	delete(set.set, i)
}
