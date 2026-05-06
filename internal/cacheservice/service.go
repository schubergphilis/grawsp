package cacheservice

import (
	"sync"

	"go.etcd.io/bbolt"
)

var (
	cache *CacheService
	once  sync.Once
)

type CacheService struct {
	db *bbolt.DB
}

func Finalize() error {
	err := cache.Finalize()
	return err
}

func (c *CacheService) Finalize() error {
	err := c.db.Close()
	return err
}

func Init(path string) error {
	var err error

	once.Do(func() {
		cache = &CacheService{}
		err = cache.Init(path)
	})

	return err
}

func (c *CacheService) Init(path string) error {
	db, err := bbolt.Open(path, 0640, nil)

	if err != nil {
		return err
	}

	c.db = db
	return err
}

func Purge(orgName string) error {
	err := cache.Purge(orgName)
	return err
}

func (c *CacheService) Purge(orgName string) error {
	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		orgBucket = nil
		err := tx.DeleteBucket([]byte(orgName))

		if err != nil {
			return err
		}

		return nil
	})

	return err
}
