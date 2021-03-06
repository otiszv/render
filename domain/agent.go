package domain

import (
	"github.com/otiszv/render/domain/common"
	"github.com/otiszv/render/jenkinsfile"
	"fmt"
)

func ValidateAgent(agent interface{}) error {
	if agent == nil {
		return nil
	}

	_, isString := agent.(string)
	_, isMapString := agent.(map[interface{}]interface{})
	_, isMapInterface := agent.(map[string]interface{})
	_, isAgent := agent.(jenkinsfile.Agent)

	if isString || isMapString || isMapInterface || isAgent {
		return nil
	}
	return common.NewTemplateDefinitionError(fmt.Sprintf("agent should be string or map[string]interface{} or map[interface{}]interface{} or Agent struct, but %T", agent), nil)
}
