package caches

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"

	// driver
	"github.com/radozd/goutils/files"
	_ "modernc.org/sqlite"
)

type PermanentCache struct {
	*sql.DB
	dbpath string
	lock   sync.RWMutex
}

func NewPermanentCache(fname string) *PermanentCache {
	return &PermanentCache{
		dbpath: fname,
	}
}

func (c *PermanentCache) Open() error {
	log.Println("DB: using " + c.dbpath)

	create := !files.Exists(c.dbpath)
	var err error
	if c.DB, err = sql.Open("sqlite", c.dbpath); err != nil {
		return err
	}

	if create {
		c.DB.Exec("PRAGMA journal_mode=WAL;")

		if _, err := c.Exec(
			`CREATE TABLE IF NOT EXISTS cache (
				key     VARCHAR (255) NOT NULL PRIMARY KEY,
				created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
				comment VARCHAR (255) NOT NULL,
				kind    VARCHAR (32) NOT NULL,
				value   BLOB
			);`); err != nil {
			return err
		}
	}
	return nil
}

func (c *PermanentCache) Close() {
	c.DB.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	c.DB.Close()
}

func (c *PermanentCache) Vacuum() {
	log.Println("DB: vacuum " + c.dbpath)
	c.Exec("VACUUM;")
}

func (c *PermanentCache) Contains(key string) (bool, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	rows, err := c.DB.Query("SELECT kind, value FROM cache WHERE key=? AND value IS NOT NULL", key)
	if err != nil {
		return false, err
	}
	contains := rows.Next()
	rows.Close()

	return contains, nil
}

func (c *PermanentCache) Remove(key string) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	res, err := c.DB.Exec("DELETE FROM cache WHERE key=?", key)
	if err != nil {
		return false, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return true, err
	}
	return rows > 0, nil
}

// Put: insert or overwrite key
func (c *PermanentCache) Put(key string, comment string, data []byte, compress string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	var value []byte
	switch compress {
	case "zlib":
		if value, err = ZlibPack(data); err != nil {
			return err
		}
	case "zstd":
		if value, err = ZstdPack(data); err != nil {
			return err
		}
	case "":
		value = data
	default:
		panic("unknown compression type")
	}

	_, err = c.DB.Exec("INSERT OR REPLACE INTO cache(key, comment, kind, value) VALUES(?,?,?,?)",
		key, comment, compress, value)
	return err
}

func (c *PermanentCache) Get(key string) ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	rows, err := c.DB.Query("SELECT kind, value FROM cache WHERE key=?", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var kind string
		var value []byte
		if err = rows.Scan(&kind, &value); err != nil {
			return nil, err
		}
		switch kind {
		case "zlib":
			return ZlibUnpack(value)
		case "zstd":
			return ZstdUnpack(value)
		case "":
			return value, nil
		default:
			return nil, errors.New("unknown compression type")
		}
	}
	return nil, nil
}

func (c *PermanentCache) GetComment(key string) (string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var comment string
	rows, err := c.DB.Query("SELECT comment FROM cache WHERE key=?", key)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&comment)
	}
	return comment, err
}

func (c *PermanentCache) PutFile(path string, baseNameAsKey bool, compress bool) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var key string
	if baseNameAsKey {
		key = filepath.Base(path)
	} else {
		key = path
	}
	return c.Put(key, "", buf, "zstd")
}

func (c *PermanentCache) ListKeys() ([]string, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	rows, err := c.DB.Query("SELECT key FROM cache")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make([]string, 0)
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (c *PermanentCache) MemCache() (map[string][]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cache := make(map[string][]byte)
	rows, err := c.DB.Query("SELECT key, kind, value FROM cache")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, kind string
		var value []byte
		if err = rows.Scan(&key, &kind, &value); err == nil {
			switch kind {
			case "zlib":
				cache[key], err = ZlibUnpack(value)
			case "zstd":
				cache[key], err = ZstdUnpack(value)
			case "":
				cache[key] = value
			default:
				err = errors.New("unknown compression type")
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return cache, nil
}
