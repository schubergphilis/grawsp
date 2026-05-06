package awsservice

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/errors"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsSession struct {
	Data          *model.Session
	ssoClient     *sso.Client
	ssooidcClient *ssooidc.Client
}

func NewAwsSession() *AwsSession {
	session := &AwsSession{}
	session.Data = &model.Session{
		AccessToken:           "",
		AccessTokenExpiresAt:  time.Time{},
		ClientId:              "",
		ClientSecret:          "",
		ClientSecretExpiresAt: time.Time{},
		DeviceCode:            "",
		DeviceExpiresAt:       time.Time{},
		VerificationUrl:       "",
	}

	return session
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

func (s *AwsSession) LoadFromCache(orgName string) error {
	sessionData, err := cacheservice.GetSession(orgName)

	if err != nil {
		if _, ok := err.(*errors.CacheMissError); ok {
			return nil
		} else {
			return err
		}
	}

	if sessionData != nil {
		s.Data = sessionData
	}

	return nil
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

func (s *AwsSession) SaveToCache(orgName string) error {
	err := cacheservice.PutSession(orgName, s.Data)
	return err
}
