package awsservice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
)

type Org struct {
	Name        string              `json:"name"`
	Region      string              `json:"region"`
	StartUrl    string              `json:"start_url"`
	Roles       map[string][]string `json:"roles"`
	DefaultRole string              `json:"default_role"`
}

type OrgSession struct {
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	ClientId              string    `json:"client_id"`
	ClientSecret          string    `json:"client_secret"`
	ClientSecretExpiresAt time.Time `json:"client_secret_expires_at"`
	DeviceCode            string    `json:"device_code"`
	DeviceExpiresAt       time.Time `json:"device_expires_at"`
	VerificationUrl       string    `json:"verification_url"`

	ssoClient     *sso.Client     `json:"-"`
	ssooidcClient *ssooidc.Client `json:"-"`
}

func (s *OrgSession) Authenticate(clientName, startUrl string) error {
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

func (s *OrgSession) authorizeDevice(startUrl string) error {
	authorization, err := s.ssooidcClient.StartDeviceAuthorization(
		context.TODO(),
		&ssooidc.StartDeviceAuthorizationInput{
			ClientId:     aws.String(s.ClientId),
			ClientSecret: aws.String(s.ClientSecret),
			StartUrl:     aws.String(startUrl),
		},
	)

	if err != nil {
		return err
	}

	s.DeviceCode = *authorization.DeviceCode
	s.DeviceExpiresAt = time.Now().Add(time.Duration(authorization.ExpiresIn) * time.Second)
	s.VerificationUrl = *authorization.VerificationUriComplete

	return nil
}

func (s *OrgSession) CreateAccessToken() error {
	tokens, err := s.ssooidcClient.CreateToken(
		context.TODO(),
		&ssooidc.CreateTokenInput{
			ClientId:     aws.String(s.ClientId),
			ClientSecret: aws.String(s.ClientSecret),
			DeviceCode:   aws.String(s.DeviceCode),
			GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		},
	)

	if err != nil {
		return err
	}

	s.AccessToken = *tokens.AccessToken
	s.AccessTokenExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)

	return nil
}

func (s *OrgSession) Init(awsConfig aws.Config) error {
	s.ssoClient = sso.NewFromConfig(awsConfig)
	s.ssooidcClient = ssooidc.NewFromConfig(awsConfig)

	return nil
}

func (s *OrgSession) IsAccessTokenValid() bool {
	if s.AccessToken != "" && s.AccessTokenExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *OrgSession) isClientRegistered() bool {
	if s.ClientId != "" && s.ClientSecret != "" && s.ClientSecretExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *OrgSession) isDeviceAuthorized() bool {
	if s.DeviceCode != "" && s.VerificationUrl != "" && s.DeviceExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (s *OrgSession) ListAccounts() ([]*Account, error) {
	var nextToken *string
	var accounts []*Account

	for {
		page, err := s.ssoClient.ListAccounts(context.TODO(), &sso.ListAccountsInput{
			AccessToken: aws.String(s.AccessToken),
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

func (s *OrgSession) registerClient(name string) error {
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

	s.ClientId = *registration.ClientId
	s.ClientSecret = *registration.ClientSecret
	s.ClientSecretExpiresAt = time.Unix(registration.ClientSecretExpiresAt, 0)

	return nil
}
