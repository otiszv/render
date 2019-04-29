package arguments

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/otiszv/render/domain/common"
)

type ArgSections []ArgSection

func (argSections ArgSections) AllArgItems() []ArgItem {
	allArgItems := []ArgItem{}
	for _, section := range argSections {
		allArgItems = append(allArgItems, section.Items...)
	}

	return allArgItems
}

type ArgSection struct {
	DisplayName common.MulitLangValue `json:"displayName" mapstructure:"displayName" yaml:"displayName"`
	Items       []ArgItem             `json:"items"`
}

type ArgItem struct {
	Name        string             `json:"name"`
	Binding     []string           `json:"binding"`
	Schema      *ArgItemSchema     `json:"schema"`
	Required    bool               `json:"required"`
	Default     interface{}        `json:"default"`
	Validation  *ArgItemValidation `json:"validation"`
	DisplayInfo *ArgDisplayInfo    `json:"display" mapstructure:"display" yaml:"display"`
	Relation    *common.Relation   `json:"relation" mapstructure:"relation" yaml:"relation"`
}
type ArgDisplayInfo struct {
	Type        string                 `json:"type"`
	Name        common.MulitLangValue  `json:"name"`
	Args        map[string]interface{} `json:"args"`
	Description common.MulitLangValue  `json:"description"`
}

func (arg *ArgItem) GetImplementor() IArgItem {
	implementorNew, _ := ArgItemImplementors[arg.Schema.Type]
	implementor := implementorNew(*arg)
	return implementor
}

func (arg *ArgItem) ValidateDefinition() error {
	//TODO
	if strings.TrimSpace(arg.Name) == "" {
		return common.NewTemplateDefinitionError("name should not be empty", nil)
	}

	if arg.Schema == nil {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.schema is required", arg.Name), nil)
	}
	if _, ok := ArgItemImplementors[arg.Schema.Type]; !ok {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.schema.type=%s is not support now", arg.Name, arg.Schema.Type), nil)
	}

	if arg.DisplayInfo == nil {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display is required", arg.Name), nil)
	}
	if arg.DisplayInfo.Type == "" {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type is required", arg.Name), nil)
	}
	if arg.DisplayInfo.Name.ZH_CN == "" {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.Name.zh-CN is required", arg.Name), nil)
	}
	if arg.DisplayInfo.Name.EN == "" {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.Name.en is required", arg.Name), nil)
	}

	implementor := arg.GetImplementor()
	return implementor.ValidateDefinition()
}

func (arg *ArgItem) ValidateValue(value interface{}) error {

	// required
	if arg.Required && value == nil {
		return common.NewValidateError(fmt.Sprintf("%s is required", arg.Name), nil)
	}

	if value == nil && arg.Required == false {
		return nil
	}

	implementor := arg.GetImplementor()
	return implementor.ValidateValue(value)
}

func (arg *ArgItem) IsMeaningful(argumentsValues map[string]interface{}) bool {
	meaningful := arg.Relation.IsMathcShowAction(argumentsValues)
	fmt.Printf("arg `%s` meaningful = %t \n", arg.Name, meaningful)
	return meaningful
}

// GetValue: get value from provider value , you should ValidateValue at first.
func (arg *ArgItem) GetValue(value interface{}) interface{} {
	return arg.GetImplementor().GetValue(value)
}

type ArgItemValidation struct {
	Pattern   string `json:"pattern"`
	MaxLength int    `json:"maxLength"`
}

type ArgItemSchema struct {
	Type  string             `json:"type"`
	Items *ArgItemSchemaItem `json:"items"`
}

type ArgItemSchemaItem struct {
	Type string `json:"type"`
}

type IArgItem interface {
	ValidateDefinition() error
	ValidateValue(value interface{}) error
	GetValue(value interface{}) interface{}
}

var ArgItemImplementors = map[string]func(data ArgItem) IArgItem{}

func init() {
	ArgItemImplementors[ArgValueTypeEnum.String] = func(data ArgItem) IArgItem {
		item := ArgItemString(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.Boolean] = func(data ArgItem) IArgItem {
		item := ArgItemBoolean(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.Object] = func(data ArgItem) IArgItem {
		item := ArgItemObject(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.Int] = func(data ArgItem) IArgItem {
		item := ArgItemInt(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.ImageRepository_Devops_IO] = func(data ArgItem) IArgItem {
		item := ArgItemRepositoryMix(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.K8SEnv_Alauda_IO] = func(data ArgItem) IArgItem {
		item := ArgItemK8sEnv(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.NewK8sContainerMix] = func(data ArgItem) IArgItem {
		item := ArgItemContainerMix(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.V1NewK8sContainerMix] = func(data ArgItem) IArgItem {
		item := ArgItemV1ContainerMix(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.Array] = func(data ArgItem) IArgItem {
		item := ArgItemArray(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.CodeRepositoryMix] = func(data ArgItem) IArgItem {
		item := ArgItemCodeRepositoryMix(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.ToolBinding] = func(data ArgItem) IArgItem {
		item := ArgItemToolBinding(data)
		return &item
	}

	ArgItemImplementors[ArgValueTypeEnum.DockerImageRepositoryMix] = func(data ArgItem) IArgItem {
		item := ArgItemDockerImageRepositoryMix(data)
		return &item
	}
}

var ArgValueTypeEnum = struct {
	String  string
	Boolean string
	Object  string
	Int     string
	Array   string

	K8SEnv_Alauda_IO     string
	V1NewK8sContainerMix string
	NewK8sContainerMix   string

	ImageRepository_Devops_IO string
	CodeRepositoryMix         string
	DockerImageRepositoryMix  string
	ToolBinding               string
}{
	Array:   "array",
	String:  "string",
	Boolean: "boolean",
	Object:  "object",
	Int:     "int",

	ImageRepository_Devops_IO: "windcloud/imagerepositorymix",
	K8SEnv_Alauda_IO:          "windcloud/k8senv",
	NewK8sContainerMix:        "windcloud/newk8scontainermix",
	V1NewK8sContainerMix:      "windcloud/v1newk8scontainermix",
	CodeRepositoryMix:         "windcloud/coderepositorymix",
	ToolBinding:               "windcloud/toolbinding",
	DockerImageRepositoryMix:  "windcloud/dockerimagerepositorymix",
}

type ArgValueType string

type ArgItemString ArgItem

const (
	DisplayInfoIntegration string = "windcloud/integration"
)

func (stringArg *ArgItemString) ValidateDefinition() error {
	if stringArg.DisplayInfo.Type == DisplayInfoIntegration {
		// need key types
		if _, ok := stringArg.DisplayInfo.Args["types"]; !ok {
			return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.args require key \"types\"", stringArg.Name), nil)
		}
	}
	return nil
}

func (stringArg *ArgItemString) ValidateValue(value interface{}) error {
	var (
		v  = ""
		ok = true
	)

	// value type validate
	if v, ok = value.(string); !ok {
		return common.NewValidateError(fmt.Sprintf("%s should be string", stringArg.Name), map[string]interface{}{
			"RecievedValue": value,
			"RecievedType":  fmt.Sprintf("%T", value),
		})
	}

	// value validate
	if stringArg.Validation != nil {
		if stringArg.Validation.MaxLength > 0 {
			if len(v) > stringArg.Validation.MaxLength {
				return common.NewValidateError(fmt.Sprintf("%s is to long, length should be less than %d",
					stringArg.Name, stringArg.Validation.MaxLength), map[string]interface{}{
					"RecievedValue": value,
					"MaxLength":     stringArg.Validation.MaxLength,
				})
			}
		}

		if stringArg.Validation.Pattern != "" {
			if matched, _ := regexp.MatchString(stringArg.Validation.Pattern, v); !matched {
				return common.NewValidateError(fmt.Sprintf("%s's value %v is not match the pattern %s",
					stringArg.Name, v, stringArg.Validation.Pattern), map[string]interface{}{
					"RecievedValue": value,
					"Pattern":       stringArg.Validation.Pattern,
				})
			}
		}
	}

	return nil
}

func (stringArg *ArgItemString) GetValue(value interface{}) interface{} {
	if value == nil {
		if stringArg.Default == nil {
			return ""
		}
		return stringArg.Default
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

type ArgItemBoolean ArgItem

func (arg *ArgItemBoolean) ValidateDefinition() error {
	return nil
}

func (arg *ArgItemBoolean) ValidateValue(value interface{}) error {
	var (
		ok = true
	)

	// value type validate
	if _, ok = value.(bool); !ok && !isStringLiteralBoolValue(value) {
		return common.NewValidateError(fmt.Sprintf("%s should be boolean", arg.Name), map[string]interface{}{
			"RecievedValue": value,
			"RecievedType":  fmt.Sprintf("%T", value),
		})
	}

	return nil
}

func (arg *ArgItemBoolean) GetValue(value interface{}) interface{} {
	if value == nil {
		if arg.Default == nil {
			return false
		}
		return arg.Default
	}

	if isStringLiteralBoolValue(value) {
		parsedValue, _ := strconv.ParseBool(value.(string))
		return parsedValue
	}
	return value
}

type ArgItemObject ArgItem

func (arg *ArgItemObject) ValidateDefinition() error {
	return nil
}

func (arg *ArgItemObject) ValidateValue(value interface{}) error {
	return nil
}

func (arg *ArgItemObject) GetValue(value interface{}) interface{} {
	if value == nil {
		if arg.Default == nil {
			return struct{}{}
		}
		return arg.Default
	}
	return value
}

type ArgItemInt ArgItem

func (arg *ArgItemInt) ValidateDefinition() error {
	return nil
}

func (arg *ArgItemInt) ValidateValue(value interface{}) error {
	return nil
}

func (arg *ArgItemInt) GetValue(value interface{}) interface{} {
	if value == nil {
		if arg.Default == nil {
			return 0
		}
		return arg.Default
	}
	return value
}

type ArgItemRepositoryMix ArgItem

func (arg *ArgItemRepositoryMix) ValidateDefinition() error {
	if arg.DisplayInfo.Type != ArgValueTypeEnum.ImageRepository_Devops_IO {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type should be %s", arg.Name, ArgValueTypeEnum.ImageRepository_Devops_IO), nil)
	}

	return nil
}

func (arg *ArgItemRepositoryMix) ValidateValue(_value interface{}) error {
	value := arg.GetValue(_value)
	if value == nil {
		return nil
	}
	v := map[string]interface{}{}
	ok := true

	if v, ok = value.(map[string]interface{}); !ok {
		return common.NewValidateError(fmt.Sprintf("%s's value %v is invalid , but got type %T", arg.Name, value, value), nil)
	}

	var requiredFields = []string{
		"registry",
		"repository",
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

func (arg *ArgItemRepositoryMix) GetValue(value interface{}) interface{} {
	return value
}

type ArgItemV1ContainerMix ArgItem

func (arg *ArgItemV1ContainerMix) ValidateDefinition() error {
	if arg.DisplayInfo.Type != ArgValueTypeEnum.V1NewK8sContainerMix {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type should be %s", arg.Name, ArgValueTypeEnum.V1NewK8sContainerMix), nil)
	}

	return nil
}

func (arg *ArgItemV1ContainerMix) ValidateValue(_value interface{}) error {
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
		"clusterName",
		"namespace",
		"applicationName",
		"componentName",
		"componentType",
		"containerName",
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

func (arg *ArgItemV1ContainerMix) GetValue(value interface{}) interface{} {
	return value
}

type ArgItemContainerMix ArgItem

func (arg *ArgItemContainerMix) ValidateDefinition() error {
	if arg.DisplayInfo.Type != ArgValueTypeEnum.NewK8sContainerMix {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type should be %s", arg.Name, ArgValueTypeEnum.NewK8sContainerMix), nil)
	}

	return nil
}

func (arg *ArgItemContainerMix) ValidateValue(_value interface{}) error {
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
		"clusterName",
		"serviceName",
		"containerName",
		"namespace",
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

func (arg *ArgItemContainerMix) GetValue(value interface{}) interface{} {
	return value
}

type ArgItemK8sEnv ArgItem

func (arg *ArgItemK8sEnv) ValidateDefinition() error {
	if arg.DisplayInfo.Type != ArgValueTypeEnum.K8SEnv_Alauda_IO {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.display.type should be %s", arg.Name, ArgValueTypeEnum.K8SEnv_Alauda_IO), nil)
	}

	return nil
}

func (arg *ArgItemK8sEnv) ValidateValue(_value interface{}) error {
	value := arg.GetValue(_value)
	if value == nil {
		return nil
	}
	v := []interface{}{}
	ok := true

	if v, ok = value.([]interface{}); !ok {
		return common.NewValidateError(fmt.Sprintf("argument %s(%s)'s value %v is invalid format, but got type %T",
			arg.Schema.Type, arg.Name, value, value), nil)
	}

	for _, item := range v {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return common.NewValidateError(fmt.Sprintf("argument %s(%s)'s value %v is invalid format, it should be map in array, but got type %T",
				arg.Schema.Type, arg.Name, value, value), nil)
		}

		name, nameExist := itemMap["name"]
		if !nameExist {
			return common.NewValidateError(fmt.Sprintf("[].name is required for argument %s, but get value %v", arg.Name, v), nil)
		}

		if name == "" {
			return common.NewValidateError(fmt.Sprintf("[].name shoule not be empty for argument %s", arg.Name), nil)
		}

		_, withValue := itemMap["value"]
		_, withValueFrom := itemMap["valueFrom"]
		// 不能都为true
		if withValue && withValueFrom {
			return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format", arg.Name, item), nil)
		}
		//不能都为false
		if withValue == false && withValueFrom == false {
			return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format", arg.Name, item), nil)
		}

		if withValue {
			if itemMap["value"] == "" {
				return common.NewValidateError(fmt.Sprintf("[].value shoule not be empty for argument %s", arg.Name), nil)
			}
		} else {
			valueFrom, ok := itemMap["valueFrom"].(map[string]interface{})
			if !ok {
				return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format, should contains valueFrom", arg.Name, item), nil)
			}
			ref, ok := valueFrom["configMapKeyRef"]
			if !ok {
				return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format, should contains configMapKeyRef", arg.Name, item), nil)
			}
			refValue, ok := ref.(map[string]interface{})
			if !ok {
				return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format, configMapKeyRef should be map", arg.Name, item), nil)
			}
			if fmt.Sprint(refValue["key"]) == "" || fmt.Sprint(refValue["name"]) == "" {
				return common.NewValidateError(fmt.Sprintf("argument %s's item value `%v` is invalid format, should contains configMapKeyRef", arg.Name, item), nil)
			}
		}

	}

	return nil
}

func (arg *ArgItemK8sEnv) GetValue(value interface{}) interface{} {
	return value
}

type ArgItemArray ArgItem

func (arg *ArgItemArray) ValidateDefinition() error {
	if arg.Schema.Items == nil {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.schema.items should not be nil", arg.Name), nil)
	}

	if _, ok := ArgItemImplementors[arg.Schema.Items.Type]; !ok {
		return common.NewTemplateDefinitionError(fmt.Sprintf("%s.schema.items.type=%s is not support now", arg.Name, arg.Schema.Type), nil)
	}

	itemImplementorNew, _ := ArgItemImplementors[arg.Schema.Items.Type]
	itemImplementor := itemImplementorNew(ArgItem(*arg))
	return itemImplementor.ValidateDefinition()
}

func (arg *ArgItemArray) ValidateValue(value interface{}) error {

	items, ok := value.([]interface{})
	if !ok {
		return common.NewValidateError(fmt.Sprintf("argument %s‘s value is invalid format ,it should be an array, but got type %T", arg.Name, value), nil)
	}

	itemImplementorNew, _ := ArgItemImplementors[arg.Schema.Items.Type]
	itemImplementor := itemImplementorNew(ArgItem(*arg))
	errs := common.Errors{}
	for _, item := range items {
		err := itemImplementor.ValidateValue(item)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func (arg *ArgItemArray) GetValue(value interface{}) interface{} {
	return value
}

func isStringLiteralBoolValue(value interface{}) bool {
	if strValue, ok := value.(string); ok {
		_, err := strconv.ParseBool(strValue)
		if err == nil {
			return true
		}
	}
	return false
}
