package domain

import (
	"gitlab.uaus.cn/devops/jenkinsrender/domain/arguments"
	"gitlab.uaus.cn/devops/jenkinsrender/domain/common"
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const CloneTaskTemplateArgName = "SCM"
const CloneTaskTemplateTypeName = "clone"

type TaskTemplateSpec struct {
	Engine    string              `json:"engine"`
	Agent     interface{}         `json:"agent"`
	Body      string              `json:"body"`
	Arguments []arguments.ArgItem `json:"arguments"`
}

func (spec *TaskTemplateSpec) ValidateDefinition() error {
	errs := common.Errors{}
	if strings.TrimSpace(spec.Body) == "" {
		errs = append(errs, common.NewTemplateDefinitionError(fmt.Sprintf("TaskTemplateSpec.Body should not be empty"), nil))
	}

	if err := ValidateAgent(spec.Agent); err != nil {
		errs = append(errs, err)
	}

	for _, argItem := range spec.Arguments {
		err := argItem.ValidateDefinition()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (spec *TaskTemplateSpec) ValidateValue(templateArgValues map[string]interface{}) error {
	errs := common.Errors{}

	for _, arg := range spec.Arguments {
		v, ok := templateArgValues[arg.Name]
		if ok == false {
			v = nil
		}

		if !arg.IsMeaningful(templateArgValues) {
			fmt.Printf("arg `%s` is not meaningful , skip validate value \n", arg.Name)
			continue
		}

		err := arg.ValidateValue(v)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (spec *TaskTemplateSpec) GetValues(templateArgValues map[string]interface{}) map[string]interface{} {
	values := make(map[string]interface{}, len(spec.Arguments))
	for _, arg := range spec.Arguments {
		v, _ := templateArgValues[arg.Name]
		values[arg.Name] = arg.GetValue(v)
	}

	// append system arg value
	if v, ok := templateArgValues[SystemArgKey]; ok {
		values[SystemArgKey] = v
	}

	return values
}

func (spec *TaskTemplateSpec) Render(templateArgValues map[string]interface{}) (string, error) {
	err := spec.ValidateDefinition()
	if err != nil {
		return "", err
	}

	err = spec.ValidateValue(templateArgValues)
	if err != nil {
		return "", err
	}

	values := spec.GetValues(templateArgValues)

	return spec.getRenderEngine()(values)
}

type taskTemplateRenderEngine func(templateArgValues map[string]interface{}) (string, error)

func (spec *TaskTemplateSpec) getRenderEngine() taskTemplateRenderEngine {
	switch spec.Engine {
	default: //default is gotpl
		{
			return func(templateArgValues map[string]interface{}) (string, error) {
				return spec.gotplRender(templateArgValues)
			}
		}
	}
}

// gotplRender Render steps script block by task template and values
func (spec *TaskTemplateSpec) gotplRender(values map[string]interface{}) (string, error) {
	t, err := template.New("gotpl-tasktemplate").Funcs(template.FuncMap{
		"split":   strings.Split,
		"replace": strings.Replace,
	}).Parse(spec.Body)
	if err != nil {
		fmt.Printf("parse task template script body error:%#v\n", err)
		return "", common.NewTemplateRenderError(err.Error(), err, nil)
	}

	buffer := bytes.NewBufferString("")
	err = t.Execute(buffer, values)
	if err != nil {
		fmt.Printf("parse task template script body execute error , body is \n %s , valus is \n %#v error is: %#v\n", spec.Body, values, err)
		return "", common.NewTemplateRenderError(fmt.Sprintf("parse task template script body execute error %s", err), err, nil)
	}

	return buffer.String(), nil
}
