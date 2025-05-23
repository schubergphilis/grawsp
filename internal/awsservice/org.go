package awsservice

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/schubergphilis/grawsp/internal/model"
)

type AwsOrg struct {
	Data model.Org

	Session *AwsSession
}

func (o *AwsOrg) StartSession(awsConfig aws.Config, clientName string) error {
	sessionData, err := cacheservice.GetSession(o.Data.Name)

	if err != nil {
		return err
	}

	o.Session = &AwsSession{}

	if sessionData != nil {
		o.Session.Data = *sessionData
	}

	err = o.Session.Init(awsConfig)

	if err != nil {
		return err
	}

	if o.Session.IsAccessTokenValid() {
		return nil
	}

	err = o.Session.Authenticate(clientName, o.Data.StartUrl)

	if err != nil {
		return err
	}

	err = cacheservice.PutSession(o.Data.Name, &o.Session.Data)

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

	err := cacheservice.PutSession(o.Data.Name, &o.Session.Data)

	return err
}
