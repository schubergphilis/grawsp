package awsservice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsSession struct {
	Data model.Session

	ssoClient     *sso.Client
	ssooidcClient *ssooidc.Client
}

func (s *AwsSession) Authenticate(clientName, startUrl string) error {
	if !s.isClientRegistered() {
		if err := s.registerClient(clientName); err != nil {
			return err
		}
	}

	if !s.isDeviceAuthorized() {
		if err := s.authorizeDevice(startUrl); err != nil {
			return err
		}
	}

	return nil
}

func (s *AwsSession) authorizeDevice(startUrl string) error {
	authorization, err := s.ssooidcClient.StartDeviceAuthorization(
		context.TODO(),
		&ssooidc.StartDeviceAuthorizationInput{
			ClientId:     aws.String(s.Data.ClientId),
			ClientSecret: aws.String(s.Data.ClientSecret),
			StartUrl:     aws.String(startUrl),
		},
	)

	if err != nil {
		return err
	}

	s.Data.DeviceCode = *authorization.DeviceCode
	s.Data.DeviceExpiresAt = time.Now().Add(time.Duration(authorization.ExpiresIn) * time.Second)
	s.Data.VerificationUrl = *authorization.VerificationUriComplete

	return nil
}

func (s *AwsSession) CreateAccessToken() error {
	tokens, err := s.ssooidcClient.CreateToken(
		context.TODO(),
		&ssooidc.CreateTokenInput{
			ClientId:     aws.String(s.Data.ClientId),
			ClientSecret: aws.String(s.Data.ClientSecret),
			DeviceCode:   aws.String(s.Data.DeviceCode),
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		},
	)

	if err != nil {
		return err
	}

	s.Data.AccessToken = *tokens.AccessToken
	s.Data.AccessTokenExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	return nil
}

func (s *AwsSession) Init(awsConfig aws.Config) error {
	s.ssoClient = sso.NewFromConfig(awsConfig)
	s.ssooidcClient = ssooidc.NewFromConfig(awsConfig)

	return nil
}

func (s *AwsSession) IsAccessTokenValid() bool {
	if s.Data.AccessToken != "" && s.Data.AccessTokenExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *AwsSession) isClientRegistered() bool {
	if s.Data.ClientId != "" && s.Data.ClientSecret != "" && s.Data.ClientSecretExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *AwsSession) isDeviceAuthorized() bool {
	if s.Data.DeviceCode != "" && s.Data.VerificationUrl != "" && s.Data.DeviceExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *AwsSession) ListAccounts() ([]*Account, error) {
	var nextToken *string
	var accounts []*Account

	for {
		page, err := s.ssoClient.ListAccounts(context.TODO(), &sso.ListAccountsInput{
			AccessToken: aws.String(s.Data.AccessToken),
			NextToken:   nextToken,
		})

		if err != nil {
			return nil, err
		}

		for _, account := range page.AccountList {
			accounts = append(accounts, &Account{
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

	return accounts, nil
}

func (s *AwsSession) registerClient(name string) error {
	registration, err := s.ssooidcClient.RegisterClient(
		context.TODO(),
		&ssooidc.RegisterClientInput{
			ClientName: aws.String(name),
			ClientType: aws.String("public"),
		},
	)

	if err != nil {
		return err
	}

	s.Data.ClientId = *registration.ClientId
	s.Data.ClientSecret = *registration.ClientSecret
	s.Data.ClientSecretExpiresAt = time.Unix(registration.ClientSecretExpiresAt, 0)

	return nil
}
