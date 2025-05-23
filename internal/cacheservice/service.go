package cacheservice

import (
	"encoding/json"
	"sync"

	"github.com/schubergphilis/grawsp/internal/model"
	"go.etcd.io/bbolt"
)

var (
	cache *CacheService
	once  sync.Once
)

type CacheService struct {
	db *bbolt.DB
}

func DeleteSession(orgName string) error {
	err := cache.DeleteSession(orgName)
	return err
}

func (c *CacheService) DeleteSession(orgName string) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(orgName))
		key := []byte("session")
		err := bucket.Delete(key)

		return err
	})

	return err
}

func GetSession(orgName string) (*model.Session, error) {
	session, err := cache.GetSession(orgName)
	return session, err
}

func (c *CacheService) GetSession(orgName string) (*model.Session, error) {
	var session model.Session

	err := c.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(orgName))

		if bucket == nil {
			return nil
		}

		key := []byte("session")
		value := bucket.Get(key)

		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &session)

		return err
	})

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func Finalize() {
	cache.Finalize()
}

func (c *CacheService) Finalize() {
	c.db.Close()
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

func PutSession(orgName string, session *model.Session) error {
	err := cache.PutSession(orgName, session)
	return err
}

func (c *CacheService) PutSession(orgName string, session *model.Session) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(orgName))

		if err != nil {
			return err
		}

		key := []byte("session")
		value, err := json.Marshal(session)

		if err != nil {
			return err
		}

		err = bucket.Put(key, value)

		return err
	})

	return err
}
