package layout

import (
	"encoding/binary"
	"hash/fnv"
)

// LayoutCache caches layout split results to avoid recomputation.
// It uses an LRU eviction strategy when the cache exceeds its capacity.
type LayoutCache struct {
	capacity int
	entries  map[uint64]*cacheEntry
	head     *cacheEntry // most recent
	tail     *cacheEntry // least recent
}

type cacheEntry struct {
	key  uint64
	val  []Rect
	prev *cacheEntry
	next *cacheEntry
}

// NewLayoutCache creates a new LRU layout cache with the given capacity.
func NewLayoutCache(capacity int) *LayoutCache {
	if capacity < 1 {
		capacity = 16
	}
	return &LayoutCache{
		capacity: capacity,
		entries:  make(map[uint64]*cacheEntry),
	}
}

// Get retrieves a cached layout result. Returns nil if not found.
func (c *LayoutCache) Get(layout Layout, area Rect) []Rect {
	key := cacheKey(layout, area)
	if e, ok := c.entries[key]; ok {
		c.moveToFront(e)
		result := make([]Rect, len(e.val))
		copy(result, e.val)
		return result
	}
	return nil
}

// Insert stores a layout result in the cache.
func (c *LayoutCache) Insert(layout Layout, area Rect, rects []Rect) {
	key := cacheKey(layout, area)

	// If already exists, update and move to front
	if e, ok := c.entries[key]; ok {
		e.val = make([]Rect, len(rects))
		copy(e.val, rects)
		c.moveToFront(e)
		return
	}

	// Evict LRU if at capacity
	for len(c.entries) >= c.capacity {
		c.evict()
	}

	val := make([]Rect, len(rects))
	copy(val, rects)
	e := &cacheEntry{key: key, val: val}
	c.entries[key] = e
	c.pushFront(e)
}

// Clear removes all entries from the cache.
func (c *LayoutCache) Clear() {
	c.entries = make(map[uint64]*cacheEntry)
	c.head = nil
	c.tail = nil
}

// Len returns the number of cached entries.
func (c *LayoutCache) Len() int {
	return len(c.entries)
}

// SplitWithCache splits the area using the layout, utilizing the cache if available.
func SplitWithCache(cache *LayoutCache, layout Layout, area Rect) []Rect {
	if cache != nil {
		if cached := cache.Get(layout, area); cached != nil {
			return cached
		}
		result := layout.Split(area)
		cache.Insert(layout, area, result)
		return result
	}
	return layout.Split(area)
}

func (c *LayoutCache) moveToFront(e *cacheEntry) {
	if e == c.head {
		return
	}
	c.remove(e)
	c.pushFront(e)
}

func (c *LayoutCache) remove(e *cacheEntry) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
	e.prev = nil
	e.next = nil
}

func (c *LayoutCache) pushFront(e *cacheEntry) {
	e.next = c.head
	e.prev = nil
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

func (c *LayoutCache) evict() {
	if c.tail == nil {
		return
	}
	delete(c.entries, c.tail.key)
	c.remove(c.tail)
}

// cacheKey generates a hash key from layout parameters and area.
func cacheKey(l Layout, area Rect) uint64 {
	h := fnv.New64a()
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(area.X))
	h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(area.Y))
	h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(area.Width))
	h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(area.Height))
	h.Write(buf[:])
	h.Write([]byte{byte(l.Direction), byte(l.Flex)})
	binary.LittleEndian.PutUint64(buf[:], uint64(l.Spacing))
	h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(l.Margin.Horizontal))
	h.Write(buf[:])
	binary.LittleEndian.PutUint64(buf[:], uint64(l.Margin.Vertical))
	h.Write(buf[:])
	for _, c := range l.Constraints {
		binary.LittleEndian.PutUint64(buf[:], uint64(c.Type))
		h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(c.Value))
		h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(c.Numerator))
		h.Write(buf[:])
		binary.LittleEndian.PutUint64(buf[:], uint64(c.Denominator))
		h.Write(buf[:])
	}
	return h.Sum64()
}
