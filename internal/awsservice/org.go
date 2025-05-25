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
	Accounts  []*model.Account
	Data      *model.Org
	Session   *AwsSession
	ssoClient *sso.Client
}

func NewAwsOrg(name, region, startUrl string) *AwsOrg {
	org := &AwsOrg{}
	org.Accounts = make([]*model.Account, 0)
	org.Session = NewAwsSession()
	org.Data = &model.Org{
		DefaultRole: "default",
		Name:        name,
		Region:      region,
		Roles:       make(map[string][]string),
		StartUrl:    startUrl,
	}

	return org
}

func (o *AwsOrg) Init() error {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(o.Data.Region))

	if err != nil {
		return err
	}

	o.ssoClient = sso.NewFromConfig(awsConfig)

	o.Session.Init(awsConfig)
	return nil
}

func (o *AwsOrg) LoadFromCache() error {
	sessionData, err := cacheservice.GetSession(o.Data.Name)

	if err != nil {
		return err
	}

	if sessionData != nil {
		o.Session.Data = sessionData
	}

	accounts, err := cacheservice.GetAccounts(o.Data.Name)

	if err != nil {
		return err
	}

	o.Accounts = accounts

	return nil
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

func (o *AwsOrg) SyncAccounts() error {
	var nextToken *string
	var accounts []*model.Account

	for {
		page, err := o.ssoClient.ListAccounts(context.TODO(), &sso.ListAccountsInput{
			AccessToken: aws.String(o.Session.Data.AccessToken),
			NextToken:   nextToken,
		})

		if err != nil {
			return err
		}

		for _, account := range page.AccountList {
			accounts = append(accounts, &model.Account{
				Email: *account.EmailAddress,
				ID:    *account.AccountId,
				Name:  *account.AccountName,
			})
		}

		nextToken = page.NextToken

		if page.NextToken == nil {
			break
		}
	}

	o.Accounts = accounts

	err := cacheservice.PutAccounts(o.Data.Name, accounts)

	return err
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
