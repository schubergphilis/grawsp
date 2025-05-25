package command

import (
	"fmt"

	"github.com/schubergphilis/grawsp/internal/awsservice"
	"github.com/spf13/viper"
)

func NewAwsOrgFromConfig(name string) (*awsservice.AwsOrg, error) {
	orgKey := fmt.Sprintf("orgs.%s", name)

	if !viper.IsSet(orgKey) {
		return nil, fmt.Errorf("organization not found: %s", name)
	}

	regionKey := fmt.Sprintf("%s.region", orgKey)
	region := ""

	if viper.IsSet(regionKey) {
		region = viper.GetString(regionKey)
	} else {
		return nil, fmt.Errorf("region was not providedf for org %s", name)
	}

	startUrlKey := fmt.Sprintf("%s.start_url", orgKey)
	startUrl := ""

	if viper.IsSet(startUrlKey) {
		startUrl = viper.GetString(startUrlKey)
	} else {
		return nil, fmt.Errorf("start_url was not providedf for org %s", name)
	}

	org := awsservice.NewAwsOrg(name, region, startUrl)
	err := org.Init()

	if err != nil {
		return nil, err
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
	}

	err = org.LoadFromCache()

	if err != nil {
		return nil, err
	}

	return org, nil
}
