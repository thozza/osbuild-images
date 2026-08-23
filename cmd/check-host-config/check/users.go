package check

import (
	"fmt"
	"log"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type UsersCheckParams struct {
	Users []string `yaml:"users"`
}

func usersFromYAML(node *yaml.Node) (CheckParams, error) {
	var p UsersCheckParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if len(p.Users) == 0 {
		return nil, fmt.Errorf("'users' list must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name: "users",
		},
		Func:       usersCheck,
		FromBuildConfig: usersFromConfig,
		FromYAML:        usersFromYAML,
	})
}

func usersFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	users := c.Blueprint.Customizations.User
	if len(users) == 0 {
		return nil, nil
	}
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}
	return UsersCheckParams{Users: names}, nil
}

func usersCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(UsersCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, user := range p.Users {
		stdout, _, _, err := ExecString("id", user)
		if err != nil {
			return Fail("user does not exist:", user)
		}
		log.Printf("User %s exists: %s\n", user, stdout)
	}

	return Pass()
}
