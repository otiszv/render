package arguments

import (
	"github.com/otiszv/render/domain/common"
	"encoding/json"
	"fmt"
)

// ArgItemDockerImageRepositoryMix code repository argument type
type ArgItemDockerImageRepositoryMix ArgItem

// ValidateDefinition valiation for code repository definition
func (arg *ArgItemDockerImageRepositoryMix) ValidateDefinition() error {
	targetType := ArgValueTypeEnum.DockerImageRepositoryMix
	if arg.DisplayInfo.Type != targetType {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type should be %s", arg.Name, targetType), nil)
	}

	return nil
}

// ValidateValue validate value
func (arg *ArgItemDockerImageRepositoryMix) ValidateValue(_value interface{}) error {
	value := arg.GetValue(_value)
	if value == nil {
		return nil
	}
	v := map[string]interface{}{}
	ok := true

	if v, ok = value.(map[string]interface{}); !ok {
		return common.NewValidateError(fmt.Sprintf("argument %s(%s)'s value %v is invalid format, got type %T", arg.Schema.Type, arg.Name, value, value), nil)
	}

	var requiredFields = []string{
		"repositoryPath",
		"credentialId",
		"tag",
	}

	for _, field := range requiredFields {
		var fieldValue interface{}
		if fieldValue, ok = v[field]; !ok {
			return common.NewValidateError(fmt.Sprintf("%s is required for argument %s, but get value %v", field, arg.Name, v), nil)
		}

		if _, ok = fieldValue.(string); !ok {
			return common.NewValidateError(fmt.Sprintf("argument %s.%s's value %v is invalid format, it should be string, but got type %T",
				arg.Name, field, fieldValue, fieldValue), nil)
		}
	}

	return nil
}

// GetValue get value of code repository
func (arg *ArgItemDockerImageRepositoryMix) GetValue(value interface{}) (result interface{}) {
	if textVal, ok := value.(string); ok {
		json.Unmarshal(([]byte)(textVal), &result)
	}
	return
}
