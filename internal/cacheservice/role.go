package cacheservice

import (
	"encoding/json"
	"fmt"

	"github.com/schubergphilis/grawsp/internal/model"
	"go.etcd.io/bbolt"
)

func GetRoles(orgName, accountName string) ([]*model.Role, error) {
	roles, err := cache.GetRoles(orgName, accountName)

	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (c *CacheService) GetRoles(orgName, accountName string) ([]*model.Role, error) {
	roles := make([]*model.Role, 0)

	err := c.db.View(func(tx *bbolt.Tx) error {
		orgBucket := tx.Bucket([]byte(orgName))

		if orgBucket == nil {
			return nil
		}

		accountBucket := orgBucket.Bucket([]byte(accountName))

		if accountBucket == nil {
			return nil
		}

		err := accountBucket.ForEachBucket(func(name []byte) error {
			roleName := string(name)
			role, err := c.GetRole(orgName, accountName, roleName)

			if err != nil {
				return err
			}

			roles = append(roles, role)

			return err
		})

		return err
	})

	if err != nil {
		return nil, err
	}

	return roles, nil
}

func GetRole(orgName, accountName, roleName string) (*model.Role, error) {
	role, err := cache.GetRole(orgName, accountName, roleName)

	if err != nil {
		return nil, err
	}

	return role, nil
}

func (c *CacheService) GetRole(orgName, accountName, roleName string) (*model.Role, error) {
	var role model.Role

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

		key := []byte("data")
		value := roleBucket.Get(key)

		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &role)

		return err
	})

	if err != nil {
		return nil, err
	}

	return &role, nil
}

func PutRole(orgName, accountName string, role *model.Role) error {
	err := cache.PutRole(orgName, accountName, role)
	return err
}

func (c *CacheService) PutRole(orgName, accountName string, role *model.Role) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		orgBucket, err := tx.CreateBucketIfNotExists([]byte(orgName))

		if err != nil {
			return err
		}

		accountBucket, err := orgBucket.CreateBucketIfNotExists([]byte(accountName))

		if err != nil {
			return err
		}

		roleBucket, err := accountBucket.CreateBucketIfNotExists([]byte(role.Name))

		if err != nil {
			return err
		}

		key := []byte("data")
		value, err := json.Marshal(role)

		if err != nil {
			return err
		}

		err = roleBucket.Put(key, value)

		return err
	})

	return err
}

func PutRoles(orgName, accountName string, roles []*model.Role) error {
	err := cache.PutRoles(orgName, accountName, roles)
	return err
}

func (c *CacheService) PutRoles(orgName, accountName string, roles []*model.Role) error {
	for _, role := range roles {
		err := c.PutRole(orgName, accountName, role)

		if err != nil {
			return err
		}
	}

	return nil
}
