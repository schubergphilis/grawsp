package cacheservice

import (
	"encoding/json"
	"fmt"

	"github.com/schubergphilis/grawsp/internal/model"
	"go.etcd.io/bbolt"
)

func GetCredential(orgName, accountName, roleName string) (*model.Credential, error) {
	credential, err := cache.GetCredential(orgName, accountName, roleName)

	if err != nil {
		return nil, err
	}

	return credential, nil
}

func (c *CacheService) GetCredential(orgName, accountName, roleName string) (*model.Credential, error) {
	var credential model.Credential

	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return fmt.Errorf("org %s not found", orgName)
		}

		accountBucket := orgBucket.Bucket([]byte(accountName))

		if accountBucket == nil {
			return fmt.Errorf("account %s not found", accountName)
		}

		roleBucket := accountBucket.Bucket([]byte(roleName))

		if roleBucket == nil {
			return fmt.Errorf("role %s not found", roleName)
		}

		key := []byte("credential")
		value := roleBucket.Get(key)

		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &credential)

		return err

	})

	if err != nil {
		return nil, err
	}

	return &credential, nil
}

func PutCredential(orgName, accountName, roleName string, credential *model.Credential) error {
	err := cache.PutCredential(orgName, accountName, roleName, credential)
	return err
}

func (c *CacheService) PutCredential(orgName, accountName, roleName string, credential *model.Credential) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket, err := tx.CreateBucketIfNotExists([]byte(orgName))

		if err != nil {
			return err
		}

		accountBucket, err := orgBucket.CreateBucketIfNotExists([]byte(accountName))

		if err != nil {
			return err
		}

		roleBucket, err := accountBucket.CreateBucketIfNotExists([]byte(roleName))

		if err != nil {
			return err
		}

		key := []byte("credential")
		value, err := json.Marshal(credential)

		if err != nil {
			return err
		}

		err = roleBucket.Put(key, value)

		return err
	})

	return err
}
