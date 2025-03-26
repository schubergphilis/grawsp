package command

import (
	"encoding/json"
	"fmt"

	"github.com/schubergphilis/grawsp/internal/awsservice"
	"github.com/spf13/viper"
	"go.etcd.io/bbolt"
)

var (
	accountsBucketName   []byte = []byte("accounts")
	orgSessionBucketName []byte = []byte("org_session")
)

func GetAccountsFromCache(orgName string) ([]*awsservice.Account, error) {
	var accounts []*awsservice.Account

	err := cache.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(accountsBucketName)

		if bucket == nil {
			return nil
		}

		key := []byte(orgName)
		value := bucket.Get(key)

		if value == nil {
			return nil
		}

		err := json.Unmarshal(value, &accounts)

		return err
	})

	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func NewOrgFromConfig(name string) (*awsservice.Org, error) {
	orgKey := fmt.Sprintf("orgs.%s", name)

	if !viper.IsSet(orgKey) {
		return nil, fmt.Errorf("organization not found: %s", name)
	}

	org := &awsservice.Org{
		Name:  name,
		Roles: map[string][]string{},
	}

	regionKey := fmt.Sprintf("%s.region", orgKey)

	if viper.IsSet(regionKey) {
		org.Region = viper.GetString(regionKey)
	}

	startUrlKey := fmt.Sprintf("%s.start_url", orgKey)

	if viper.IsSet(startUrlKey) {
		org.StartUrl = viper.GetString(startUrlKey)
	}

	rolesKey := fmt.Sprintf("%s.roles", orgKey)

	if viper.IsSet(rolesKey) {
		for role := range viper.GetStringMap(rolesKey) {
			roleKey := fmt.Sprintf("%s.%s", rolesKey, role)
			org.Roles[role] = viper.GetStringSlice(roleKey)
		}
	}

	defaultRoleKey := fmt.Sprintf("%s.default_role", orgKey)

	if viper.IsSet(defaultRoleKey) {
		org.DefaultRole = viper.GetString(defaultRoleKey)
	} else {
		org.DefaultRole = "default"
	}

	return org, nil
}

func NewOrgSessionFromCache(name string) (*awsservice.OrgSession, error) {
	var err error

	session := &awsservice.OrgSession{}

	err = cache.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(orgSessionBucketName)

		if bucket == nil {
			return nil
		}

		key := []byte(name)
		value := bucket.Get(key)

		if value == nil {
			return nil
		}

		err = json.Unmarshal(value, session)

		return err
	})

	if err != nil {
		return nil, err
	}

	return session, nil
}

func WriteAccountsToCache(orgName string, accounts []*awsservice.Account) error {
	err := cache.Update(func(tx *bbolt.Tx) error {
		value, err := json.Marshal(accounts)

		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucketIfNotExists(accountsBucketName)

		if err != nil {
			return err
		}

		key := []byte(orgName)
		err = bucket.Put(key, value)

		return err
	})

	return err
}

func WriteOrgSessionToCache(name string, session *awsservice.OrgSession) error {
	err := cache.Update(func(tx *bbolt.Tx) error {
		value, err := json.Marshal(session)

		if err != nil {
			return err
		}

		bucket, err := tx.CreateBucketIfNotExists(orgSessionBucketName)

		if err != nil {
			return err
		}

		key := []byte(name)
		err = bucket.Put(key, value)

		return err
	})

	return err
}
