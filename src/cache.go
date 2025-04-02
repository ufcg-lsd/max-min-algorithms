package main

import (
	"container/list"
)

//TODO: I think we can improve readability (of the caller code) if
//		1. we add a contains method
//		2. we add a peek method (then, we make clear that get chances de LRU
//	order and peek does not)
//		3. maybe, add a bool return to the Add method to indicate an item was
// 	evicted to make room for the new item (this may help stats collection?)

// This cache is not thread-safe
type LRUCache struct {
	capacity   int64
	cache      map[string]*list.Element
	lruList    *list.List
	cacheUsage int64
}

type cacheItem struct {
	key   string
	value interface{}
	bytes int64
}

func NewLRUCache(capacity int64) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lruList:  list.New(),
	}
}

/*
Returns (V, true) if there is a value V mapped by key or (nil, false) otherwise.
After the call, the key is the most recenlty used entry in the cache.
*/
func (c *LRUCache) Get(key string) (interface{}, bool) {
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		return elem.Value.(*cacheItem).value, true
	}
	return nil, false
}

/*
Adds or update the mapping <key, value> to the cache. (and bytes? mapping?)
Returns <V, true> if there was already a value V mapped by key or <nil, false> otherwise.
After the call, the key is the most recenlty used entry in the cache.
*/
func (c *LRUCache) Add(key string, value interface{}, bytes int64) (interface{}, bool) {
	if previousElem, mapped := c.cache[key]; mapped {
		// remove element before update timestamp and move to front in lru list
		c.Remove(key)
		c.doAdd(key, value, bytes)
		return previousElem.Value.(*cacheItem).value, mapped
	}

	if bytes > c.capacity {
		return nil, false
	}

	// Until the difference between capacity and cache usage - free space - is less ,or equal, than bytes, remove lru item.
	for (c.capacity - c.cacheUsage) < bytes {
		lastElem := c.lruList.Back()
		c.Remove(lastElem.Value.(*cacheItem).key)
	}

	c.doAdd(key, value, bytes)
	return nil, false
}

func (c *LRUCache) doAdd(key string, value interface{}, bytes int64) {
	newElem := c.lruList.PushFront(&cacheItem{key: key, value: value, bytes: bytes})
	c.cacheUsage = c.cacheUsage + bytes
	c.cache[key] = newElem
}

/*
Remove an cached item using a key passed as parameter.
And decrease cache usage.
*/
func (c *LRUCache) Remove(key string) {
	if elem, ok := c.cache[key]; ok {
		c.cacheUsage = c.cacheUsage - elem.Value.(*cacheItem).bytes
		delete(c.cache, key)
		c.lruList.Remove(elem)
	}
}

/*
Remove all cache items in a list passed as parameter.

func (c *LRUCache) RemoveAll(removeList *list.List) {
	for e := removeList.Front(); e != nil; e = e.Next() {
		cache.Remove(e.Value.(string))
	}
}
*/
/*
	The size - usage capacity
*/

func (c *LRUCache) Size() int64 {
	return c.cacheUsage
}
