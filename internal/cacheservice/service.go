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

func GetAccount(orgName, accountName string) (*model.Account, error) {
	account, err := cache.GetAccount(orgName, accountName)

	if err != nil {
		return nil, err
	}

	return account, nil
}

func (c *CacheService) GetAccount(orgName, accountName string) (*model.Account, error) {
	var session model.Account

	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		bucket := orgBucket.Bucket([]byte("accounts"))

		if bucket == nil {
			return nil
		}

		key := []byte(accountName)
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

func GetAccounts(orgName string) ([]*model.Account, error) {
	accounts, err := cache.GetAccounts(orgName)

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (c *CacheService) GetAccounts(orgName string) ([]*model.Account, error) {
	var accounts []*model.Account

	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		bucket := orgBucket.Bucket([]byte("accounts"))

		if bucket == nil {
			return nil
		}

		cursor := bucket.Cursor()

		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			account := &model.Account{}
			err := json.Unmarshal(value, account)

			if err != nil {
				return err
			}

			accounts = append(accounts, account)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func GetSession(orgName string) (*model.Session, error) {
	session, err := cache.GetSession(orgName)

	if err != nil {
		return nil, err
	}

	return session, nil
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

func PutAccount(orgName string, account *model.Account) error {
	err := cache.PutAccount(orgName, account)
	return err
}

func (c *CacheService) PutAccount(orgName string, account *model.Account) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket, err := tx.CreateBucketIfNotExists([]byte(orgName))

		if err != nil {
			return err
		}

		bucket, err := orgBucket.CreateBucketIfNotExists([]byte("accounts"))

		if err != nil {
			return err
		}

		key := []byte(account.Name)
		value, err := json.Marshal(account)

		if err != nil {
			return err
		}

		err = bucket.Put(key, value)

		if err != nil {
			return err
		}

		return err
	})

	return err
}

func PutAccounts(orgName string, accounts []*model.Account) error {
	err := cache.PutAccounts(orgName, accounts)
	return err
}

func (c *CacheService) PutAccounts(orgName string, accounts []*model.Account) error {
	for _, account := range accounts {
		err := cache.PutAccount(orgName, account)

		if err != nil {
			return err
		}
	}

	return nil
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
