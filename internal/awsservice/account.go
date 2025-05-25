package awsservice

import "github.com/schubergphilis/grawsp/internal/model"

type AwsAccount struct {
	Data *model.Account
}

func NewAwsAccount(id, name, email string) *AwsAccount {
	account := &AwsAccount{
		Data: &model.Account{
			Email: email,
			ID:    id,
			Name:  name,
		},
	}

	return account
}
