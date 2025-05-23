package command

import (
	"fmt"

	"github.com/schubergphilis/grawsp/internal/awsservice"
	"github.com/schubergphilis/grawsp/internal/model"
	"github.com/spf13/viper"
)

func NewAwsOrgFromConfig(name string) (*awsservice.AwsOrg, error) {
	orgKey := fmt.Sprintf("orgs.%s", name)

	if !viper.IsSet(orgKey) {
		return nil, fmt.Errorf("organization not found: %s", name)
	}

	org := &awsservice.AwsOrg{
		Data: model.Org{
			Name:  name,
			Roles: make(map[string][]string),
		},
	}

	regionKey := fmt.Sprintf("%s.region", orgKey)

	if viper.IsSet(regionKey) {
		org.Data.Region = viper.GetString(regionKey)
	} else {
		return nil, fmt.Errorf("region was not providedf for org %s", name)
	}

	startUrlKey := fmt.Sprintf("%s.start_url", orgKey)

	if viper.IsSet(startUrlKey) {
		org.Data.StartUrl = viper.GetString(startUrlKey)
	} else {
		return nil, fmt.Errorf("start_url was not providedf for org %s", name)
	}

	rolesKey := fmt.Sprintf("%s.roles", orgKey)

	if viper.IsSet(rolesKey) {
		for role := range viper.GetStringMap(rolesKey) {
			roleKey := fmt.Sprintf("%s.%s", rolesKey, role)
			org.Data.Roles[role] = viper.GetStringSlice(roleKey)
		}
	}

	defaultRoleKey := fmt.Sprintf("%s.default_role", orgKey)

	if viper.IsSet(defaultRoleKey) {
		org.Data.DefaultRole = viper.GetString(defaultRoleKey)
	} else {
		org.Data.DefaultRole = "default"
	}

	return org, nil
}
