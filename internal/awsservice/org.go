package awsservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsOrg struct {
	Accounts  *AwsAccountCollection
	awsConfig aws.Config
	Data      *model.Org
	Session   *AwsSession
	ssoClient *sso.Client
}

func NewAwsOrg(name, region, startUrl string) *AwsOrg {
	org := &AwsOrg{}
	org.Accounts = &AwsAccountCollection{}
	org.Session = NewAwsSession()
	org.Data = &model.Org{
		DefaultRole: "default",
		Name:        name,
		Region:      region,
		Roles:       make(map[string]string),
		StartUrl:    startUrl,
	}

	return org
}

func (o *AwsOrg) Init() error {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(o.Data.Region))

	if err != nil {
		return err
	}

	o.awsConfig = awsConfig
	o.ssoClient = sso.NewFromConfig(awsConfig)

	err = o.Session.Init(awsConfig)

	if err != nil {
		return err
	}

	err = o.Accounts.Init(awsConfig)

	return err
}

func (o *AwsOrg) LoadFromCache() error {
	err := o.Session.LoadFromCache(o.Data.Name)

	if err != nil {
		return err
	}

	err = o.Accounts.LoadFromCache(o.Data.Name)

	return err
}

func (o *AwsOrg) StartSession(clientName string) error {
	if o.Session.IsAccessTokenValid() {
		return nil
	}

	err := o.Session.Authenticate(clientName, o.Data.StartUrl)

	if err != nil {
		return err
	}

	err = cacheservice.PutSession(o.Data.Name, o.Session.Data)

	return err
}

func (o *AwsOrg) SaveToCache() error {
	err := o.Accounts.SaveToCache(o.Data.Name)

	if err != nil {
		return err
	}

	err = o.Session.SaveToCache(o.Data.Name)

	return err
}

func (o *AwsOrg) SyncAccounts() error {
	o.Accounts.Clear()

	paginator := sso.NewListAccountsPaginator(o.ssoClient, &sso.ListAccountsInput{
		AccessToken: aws.String(o.Session.Data.AccessToken),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())

		if err != nil {
			return err
		}

		for _, account := range page.AccountList {
			awsAccount := NewAwsAccount(
				*account.AccountId,
				*account.AccountName,
				*account.EmailAddress,
			)

			err = awsAccount.Init(o.awsConfig)

			if err != nil {
				return err
			}

			o.Accounts.Append(awsAccount)
		}
	}

	return nil
}

func (o *AwsOrg) WaitForVerification(timeout int) error {
	timeStart := time.Now()

	for {
		err := o.Session.CreateAccessToken()

		if err == nil {
			break
		}

		var bne *types.AuthorizationPendingException

		if errors.As(err, &bne) {
			timeElapsed := time.Since(timeStart)

			if timeElapsed.Seconds() >= float64(timeout) {
				return fmt.Errorf("timeout reached")
			}

			time.Sleep(2 * time.Second)
		} else {
			return err
		}
	}

	err := cacheservice.PutSession(o.Data.Name, o.Session.Data)

	return err
}
