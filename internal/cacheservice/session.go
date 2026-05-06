package cacheservice

import (
	"encoding/json"

	"github.com/schubergphilis/grawsp/internal/errors"
	"github.com/schubergphilis/grawsp/internal/model"
	"go.etcd.io/bbolt"
)

func DeleteSession(orgName string) error {
	err := cache.DeleteSession(orgName)
	return err
}

func (c *CacheService) DeleteSession(orgName string) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		key := []byte("session")
		err := orgBucket.Delete(key)

		return err
	})

	return err
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
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return &errors.CacheMissError{
				ObjectName: orgName,
			}
		}

		key := []byte("session")
		value := orgBucket.Get(key)

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

func PutSession(orgName string, session *model.Session) error {
	err := cache.PutSession(orgName, session)
	return err
}

func (c *CacheService) PutSession(orgName string, session *model.Session) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket, err := tx.CreateBucketIfNotExists([]byte(orgName))

		if err != nil {
			return err
		}

		key := []byte("session")
		value, err := json.Marshal(session)

		if err != nil {
			return err
		}

		err = orgBucket.Put(key, value)

		return err
	})

	return err
}
