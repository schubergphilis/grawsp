package cacheservice

import (
	"encoding/json"
	"fmt"

	"github.com/schubergphilis/grawsp/internal/model"
	"go.etcd.io/bbolt"
)

func DeleteAllAccounts(orgName string) error {
	err := cache.DeleteAllAccounts(orgName)
	return err
}

func (c *CacheService) DeleteAllAccounts(orgName string) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		err := orgBucket.ForEachBucket(func(name []byte) error {
			err := orgBucket.DeleteBucket(name)
			return err
		})

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
	var account model.Account

	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		accountBucket := orgBucket.Bucket([]byte(accountName))

		if accountBucket == nil {
			return fmt.Errorf("account %s not found", accountName)
		}

		key := []byte("data")
		value := accountBucket.Get(key)

		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &account)

		return err
	})

	if err != nil {
		return nil, err
	}

	return &account, nil
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

		err := orgBucket.ForEachBucket(func(name []byte) error {
			account, err := c.GetAccount(orgName, string(name))

			if err != nil {
				return err
			}

			accounts = append(accounts, account)
			return nil
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return accounts, nil
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

		accountBucket, err := orgBucket.CreateBucketIfNotExists([]byte(account.Name))

		if err != nil {
			return err
		}

		key := []byte("data")
		value, err := json.Marshal(account)

		if err != nil {
			return err
		}

		err = accountBucket.Put(key, value)
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
