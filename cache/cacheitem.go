package cache2gowithclient

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex

	// Key and Value
	key   interface{}
	value interface{}

	// Time to live, current time - last accessed time > lifeSpan -> this key/value expires
	lifeSpan time.Duration

	// Other time related data
	createdOn   time.Time
	accessedOn  time.Time
	accessCount int
}

func NewCahceItem(key interface{}, lifeSpan time.Duration, value interface{}) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:         key,
		lifeSpan:    lifeSpan,
		createdOn:   t,
		accessedOn:  t,
		accessCount: 0,
		value:       value,
	}
}

// Some get method
func (item *CacheItem) LiftSpan() time.Duration {
	return item.lifeSpan
}

func (item *CacheItem) AccessedOn() time.Time {
	return item.accessedOn
}

func (item *CacheItem) CreatedOn() time.Time {
	return item.createdOn
}

func (item *CacheItem) AccessCount() int {
	item.RLock()
	defer item.RUnlock()
	return item.accessCount
}

func (item *CacheItem) Key() interface{} {
	return item.key
}

func (item *CacheItem) Value() interface{} {
	return item.value
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
	item.accessCount++
}
