package awsservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsFederationSignInResponse struct {
	SigninToken string `json:"SigninToken"`
}

type AwsRole struct {
	awsConfig  aws.Config
	ssoClient  *sso.Client
	Data       *model.Role
	Credential *model.Credential
}

func NewAwsRole(name string) *AwsRole {
	role := &AwsRole{
		Data: &model.Role{
			Name: name,
		},
		Credential: &model.Credential{},
	}

	return role
}

func (r *AwsRole) GetCredential(accessToken, accountId string) error {
	credential, err := r.ssoClient.GetRoleCredentials(
		context.TODO(),
		&sso.GetRoleCredentialsInput{
			AccessToken: aws.String(accessToken),
			AccountId:   aws.String(accountId),
			RoleName:    aws.String(r.Data.Name),
		},
	)

	if err != nil {
		return err
	}

	r.Credential.AccessKeyId = *credential.RoleCredentials.AccessKeyId
	r.Credential.ExpiresAt = time.UnixMilli(credential.RoleCredentials.Expiration)
	r.Credential.SecretAccessKey = *credential.RoleCredentials.SecretAccessKey
	r.Credential.SessionToken = *credential.RoleCredentials.SessionToken

	return nil
}

func (r *AwsRole) GetConsoleURL(region string) (string, error) {
	signinUrl, err := url.Parse("https://signin.aws.amazon.com/federation")

	if err != nil {
		return "", err
	}

	sessionCredential := map[string]string{
		"sessionId":    r.Credential.AccessKeyId,
		"sessionKey":   r.Credential.SecretAccessKey,
		"sessionToken": r.Credential.SessionToken,
	}

	session, err := json.Marshal(sessionCredential)

	if err != nil {
		return "", err
	}

	signinParams := url.Values{}

	signinParams.Set("Action", "getSigninToken")
	signinParams.Set("Session", string(session))

	signinUrl.RawQuery = signinParams.Encode()

	req, err := http.NewRequest(http.MethodGet, signinUrl.String(), nil)

	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Request to federated signin endpoing failed: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	var signinResponse AwsFederationSignInResponse

	if err := json.Unmarshal(body, &signinResponse); err != nil {
		return "", err
	}

	signinToken := signinResponse.SigninToken

	consoleUrl, err := url.Parse("https://signin.aws.amazon.com/federation")

	if err != nil {
		return "", err
	}

	destinationUrl := fmt.Sprintf("https://console.aws.amazon.com/?region=%s", region)

	consoleParams := url.Values{}

	consoleParams.Set("Action", "login")
	consoleParams.Set("Destination", destinationUrl)
	consoleParams.Set("SigninToken", signinToken)

	consoleUrl.RawQuery = consoleParams.Encode()

	return consoleUrl.String(), nil
}

func (r *AwsRole) Init(awsConfig aws.Config) error {
	r.ssoClient = sso.NewFromConfig(awsConfig)
	r.awsConfig = awsConfig

	return nil
}

func (r *AwsRole) IsCredentialValid() bool {
	if r.Credential.SessionToken != "" &&
		r.Credential.AccessKeyId != "" &&
		r.Credential.SecretAccessKey != "" &&
		r.Credential.ExpiresAt.After(time.Now()) {
		return true
	}

	return false
}

func (r *AwsRole) LoadCredentialFromCache(orgName, accountName string) error {
	roleName := r.Data.Name

	credentialData, err := cacheservice.GetCredential(orgName, accountName, roleName)

	if err != nil {
		return err
	}

	r.Credential = credentialData

	return nil
}

func (r *AwsRole) LoadFromCache(orgName, accountName string) error {
	roleData, err := cacheservice.GetRole(orgName, accountName, r.Data.Name)

	if err != nil {
		return err
	}

	if roleData != nil {
		r.Data = roleData
	}

	err = r.LoadCredentialFromCache(orgName, accountName)

	if err != nil {
		return err
	}

	return nil
}

func (r *AwsRole) SaveCredentialToCache(orgName, accountName string) error {
	roleName := r.Data.Name
	err := cacheservice.PutCredential(orgName, accountName, roleName, r.Credential)

	if err != nil {
		return err
	}

	return nil
}

func (r *AwsRole) SaveToCache(orgName, accountName string) error {
	err := cacheservice.PutRole(orgName, accountName, r.Data)

	if err != nil {
		return err
	}

	err = r.SaveCredentialToCache(orgName, accountName)

	if err != nil {
		return err
	}

	return nil
}

type AwsRoleCollection struct {
	roles []*AwsRole
}

func (rc *AwsRoleCollection) Append(role *AwsRole) {
	rc.roles = append(rc.roles, role)
}

func (rc *AwsRoleCollection) All() <-chan *AwsRole {
	ch := make(chan *AwsRole)

	go func() {
		for _, role := range rc.roles {
			ch <- role
		}
		close(ch)
	}()

	return ch
}

func (rc *AwsRoleCollection) Clear() {
	rc.roles = make([]*AwsRole, 0)
}

func (rc *AwsRoleCollection) FindByName(name string) *AwsRole {
	for role := range rc.All() {
		if role.Data.Name == name {
			return role
		}
	}

	return nil
}

func (rc *AwsRoleCollection) Init(awsConfig aws.Config) error {
	for role := range rc.All() {
		err := role.Init(awsConfig)

		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *AwsRoleCollection) LoadFromCache(orgName, accountName string) error {
	roles, err := cacheservice.GetRoles(orgName, accountName)

	if err != nil {
		return err
	}

	for _, role := range roles {
		awsRole := NewAwsRole(role.Name)

		err = awsRole.LoadCredentialFromCache(orgName, accountName)

		if err != nil {
			return err
		}

		rc.roles = append(rc.roles, awsRole)
	}

	return nil
}

func (rc *AwsRoleCollection) SaveToCache(orgName, accountName string) error {
	for _, role := range rc.roles {
		err := role.SaveToCache(orgName, accountName)

		if err != nil {
			return err
		}
	}

	return nil
}

func (rc *AwsRoleCollection) ToString() string {
	roleNames := make([]string, 0)

	for role := range rc.All() {
		roleNames = append(roleNames, role.Data.Name)
	}

	return strings.Join(roleNames, ",")
}
