package awsservice

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsAccount struct {
	awsConfig aws.Config
	Data      *model.Account
	Roles     *AwsRoleCollection
	ssoClient *sso.Client
}

func NewAwsAccount(id, name, email string) *AwsAccount {
	account := &AwsAccount{
		Roles: &AwsRoleCollection{},
		Data: &model.Account{
			Email: email,
			ID:    id,
			Name:  name,
		},
	}

	return account
}

func (a *AwsAccount) Init(awsConfig aws.Config) error {
	a.ssoClient = sso.NewFromConfig(awsConfig)
	a.awsConfig = awsConfig

	err := a.Roles.Init(awsConfig)

	return err
}

func (a *AwsAccount) LoadFromCache(orgName string) error {
	accountData, err := cacheservice.GetAccount(orgName, a.Data.Name)

	if err != nil {
		return err
	}

	if accountData != nil {
		a.Data = accountData
	}

	err = a.Roles.LoadFromCache(orgName, a.Data.Name)

	return err
}

func (a *AwsAccount) LoadRolesFromCache(orgName string) error {
	err := a.Roles.LoadFromCache(orgName, a.Data.Name)

	return err
}

func (a *AwsAccount) SaveToCache(orgName string) error {
	err := cacheservice.PutAccount(orgName, a.Data)

	if err != nil {
		return err
	}

	err = a.Roles.SaveToCache(orgName, a.Data.Name)

	return err
}

func (a *AwsAccount) SyncRoles(accessToken string) error {
	a.Roles.Clear()

	paginator := sso.NewListAccountRolesPaginator(a.ssoClient, &sso.ListAccountRolesInput{
		AccessToken: aws.String(accessToken),
		AccountId:   aws.String(a.Data.ID),
	})

	for paginator.HasMorePages() {
		var err error
		var page *sso.ListAccountRolesOutput

		for retries := 0; retries < 5; retries++ {
			page, err = paginator.NextPage(context.TODO())

			if err == nil {
				break
			}

			var bne *types.TooManyRequestsException

			if errors.As(err, &bne) {
				time.Sleep(1 * time.Second)
			} else {
				return err
			}
		}

		if page != nil {
			for _, role := range page.RoleList {
				r := NewAwsRole(*role.RoleName)
				r.Init(a.awsConfig)
				a.Roles.Append(r)
			}
		}
	}

	return nil
}

type AwsAccountCollection struct {
	accounts []*AwsAccount
}

func (ac *AwsAccountCollection) Append(account *AwsAccount) {
	ac.accounts = append(ac.accounts, account)
}

func (ac *AwsAccountCollection) All() <-chan *AwsAccount {
	ch := make(chan *AwsAccount)

	go func() {
		for _, account := range ac.accounts {
			ch <- account
		}
		close(ch)
	}()

	return ch
}

func (ac *AwsAccountCollection) Clear() {
	ac.accounts = make([]*AwsAccount, 0)
}

func (ac *AwsAccountCollection) FindByName(name string) *AwsAccount {
	for account := range ac.All() {
		if account.Data.Name == name {
			return account
		}
	}

	return nil
}

func (ac *AwsAccountCollection) Init(awsConfig aws.Config) error {
	for account := range ac.All() {
		err := account.Init(awsConfig)

		if err != nil {
			return err
		}
	}

	return nil
}

func (ac *AwsAccountCollection) LoadFromCache(orgName string) error {
	accounts, err := cacheservice.GetAccounts(orgName)

	if err != nil {
		return err
	}

	for _, account := range accounts {
		awsAccount := NewAwsAccount(
			account.ID,
			account.Name,
			account.Email,
		)

		err = awsAccount.LoadRolesFromCache(orgName)

		if err != nil {
			return err
		}

		ac.accounts = append(ac.accounts, awsAccount)
	}

	return nil
}

func (ac *AwsAccountCollection) SaveToCache(orgName string) error {
	err := cacheservice.DeleteAllAccounts(orgName)

	if err != nil {
		return err
	}

	for _, account := range ac.accounts {
		err = account.SaveToCache(orgName)

		if err != nil {
			return err
		}
	}

	return nil
}
