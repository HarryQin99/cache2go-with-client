package cache2gowithclient

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
)

type CacheTable struct {
	sync.RWMutex
	name  string
	items map[interface{}]*CacheItem

	// Clean up related variable
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration

	logger *log.Logger

	loadData func(key interface{}, args ...interface{}) *CacheItem
}

func (table *CacheTable) SetDataLoader(f func(interface{}, ...interface{}) *CacheItem) {
	table.Lock()
	defer table.Unlock()
	table.loadData = f
}

func (table *CacheTable) SetLogger(logger *log.Logger) {
	table.Lock()
	defer table.Unlock()
	table.logger = logger
}

func (table *CacheTable) log(v ...interface{}) {
	if table.logger == nil {
		return
	}

	table.logger.Println(v...)
}

func (table *CacheTable) expirationCheck() {
	fmt.Println("Running expiration check")
	table.Lock()
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
	if table.cleanupInterval > 0 {
		table.log("Expiration check triffered after", table.cleanupInterval, "for table", table.name)
	} else {
		table.log("Expiration check installed for table", table.name)
	}
	fmt.Println("End")
	now := time.Now()
	smallestDuration := 0 * time.Second
	for key, item := range table.items {
		fmt.Println("End")
		item.RLock()
		lifeSpan := item.lifeSpan
		accessedOn := item.accessedOn
		item.RUnlock()

		if lifeSpan == 0 {
			continue
		}

		if now.Sub(accessedOn) >= lifeSpan {
			table.deleteInternal(key)
		} else {
			if smallestDuration == 0 || lifeSpan-now.Sub(accessedOn) < smallestDuration {
				smallestDuration = lifeSpan - now.Sub(accessedOn)
			}
		}
	}

	table.cleanupInterval = smallestDuration
	if smallestDuration > 0 {
		table.cleanupTimer = time.AfterFunc(smallestDuration, func() {
			go table.expirationCheck()
		})
	}
	fmt.Println("End")
	table.Unlock()
}

// Add key/value
func (table *CacheTable) Add(key interface{}, lifeSpan time.Duration, value interface{}) *CacheItem {
	item := NewCahceItem(key, lifeSpan, value)
	fmt.Println(lifeSpan)
	table.Lock()
	table.addInternal(item)
	fmt.Println(table.items[key].key)

	return item
}

func (table *CacheTable) addInternal(item *CacheItem) {
	table.log("Adding item with key", item.key, "and value of", item.value, "with lifespan", item.lifeSpan, "to table", table.name)

	table.items[item.key] = item

	expDur := table.cleanupInterval
	table.Unlock()

	// Only check expiration while new item has a setted lifeSpan and current expDur equal to 0 or new lifeSpan smaller than current expDur
	if item.lifeSpan > 0 && (expDur == 0 || item.lifeSpan < expDur) {
		table.expirationCheck()
	}
}

func (table *CacheTable) Delete(key interface{}) (*CacheItem, error) {
	table.Lock()
	defer table.Unlock()

	item, err := table.deleteInternal(key)
	table.Unlock()
	return item, err
}

func (table *CacheTable) deleteInternal(key interface{}) (*CacheItem, error) {
	r, ok := table.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	table.Unlock()
	table.Lock()
	table.log("Deleting item with key", key, "created on", r.createdOn, "and hit", r.accessCount, "times from table", table.name)
	delete(table.items, key)
	return r, nil

}

func (table *CacheTable) Exists(key interface{}) bool {
	table.RLock()
	defer table.RUnlock()
	_, ok := table.items[key]

	return ok
}

func (table *CacheTable) NotFoundAdd(key interface{}, lifeSpan time.Duration, value interface{}) bool {
	table.Lock()

	if _, ok := table.items[key]; ok {
		table.Unlock()
		return false
	}

	item := NewCahceItem(key, lifeSpan, value)
	table.addInternal(item)

	return true
}

func (table *CacheTable) Value(key interface{}, args ...interface{}) (*CacheItem, error) {
	table.RLock()
	r, ok := table.items[key]
	loadData := table.loadData
	table.RUnlock()

	if ok {
		r.KeepAlive()
		return r, nil
	}
	if loadData != nil {
		item := loadData(key, args...)
		if item != nil {
			table.Add(key, item.lifeSpan, item.value)
			return item, nil
		}

		return nil, ErrKeyNotFound
	}

	return nil, ErrKeyNotFound
}

func (table *CacheTable) Flush() {
	table.Lock()
	defer table.Unlock()

	table.log("Flushing table", table.name)
	table.cleanupInterval = 0
	if table.cleanupTimer != nil {
		table.cleanupTimer.Stop()
	}
}

// Sort cacheItem based on access count
type CacheItemPair struct {
	Key         interface{}
	AccessCount int
}

type CacheItemPairList []CacheItemPair

func (p CacheItemPairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p CacheItemPairList) Len() int {
	return len(p)
}

func (p CacheItemPairList) Less(i, j int) bool {
	return p[i].AccessCount > p[j].AccessCount
}

func (table *CacheTable) MostAccessed(count int) []*CacheItem {
	table.Lock()
	defer table.Unlock()

	// Get the sorted item key list
	p := make(CacheItemPairList, len(table.items))
	i := 0
	for k, v := range table.items {
		p[i] = CacheItemPair{
			Key:         k,
			AccessCount: v.accessCount,
		}
		i++
	}
	sort.Sort(p)

	var r []*CacheItem
	c := 0
	for _, v := range p {
		if c > count {
			break
		}

		item, ok := table.items[v.Key]
		if ok {
			r = append(r, item)
		}
		c++
	}

	return r
}
