package lru

import "container/list"

type Cache struct {
	maxBytes  int64      // 允许使用的最大内存
	nbytes    int64      // 当前已使用的内存
	ll        *list.List // 某条记录被移除时的回调函数，可以为 nil
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct {
	key   string
	value Value
}

// 返回值所占用的内存大小
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes, //等于0不remove
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找主要有 2 个步骤，第一步是从字典中找到对应的双向链表的节点，第二步，将该节点移动到队尾。
// 链表中的节点 ele 移动到队尾（双向链表作为队列，队首队尾是相对的，在这里约定 front 为队尾）
// root(哨兵) <-> element1 <-> element2 <-> element3
//
//	 ↑            	↑
//	root           front
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 缓存淘汰。即移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if eleV, ok := c.Get(key); ok {
		c.nbytes += int64(value.Len() - eleV.Len())
		eleV = value
	} else {
		//Head -> r -> ele -> Tail
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key) + value.Len())
	}
	// 添加一个新的元素（或更新现有元素的大小）后，c.nbytes可能会超出 maxBytes 的限制，
	// 需要多次逐步移除旧数据直到总大小恢复到 maxBytes 以内
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
