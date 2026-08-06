package cache

import "sync"

// Cache хранит API-ключи в памяти.
//
// Все проверки авторизации будут
// выполняться через этот объект,
// без обращения к PostgreSQL.
type Cache struct {
	mu sync.RWMutex

	keys map[string]APIKey
}

// New создаёт пустой Runtime Cache.
func New() *Cache {

	return &Cache{

		keys: make(map[string]APIKey),
	}
}

// Get возвращает ключ.
func (c *Cache) Get(
	key string,
) (APIKey, bool) {

	c.mu.RLock()
	defer c.mu.RUnlock()

	apiKey, ok := c.keys[key]

	return apiKey, ok
}

// Replace полностью заменяет содержимое кэша.
//
// Используется после очередной загрузки
// из PostgreSQL.
func (c *Cache) Replace(
	keys map[string]APIKey,
) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.keys = keys
}

// Size возвращает количество
// загруженных ключей.
func (c *Cache) Size() int {

	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.keys)
}
