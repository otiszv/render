package common

import (
	"fmt"
)

type MulitLangValue struct {
	ZH_CN string `json:"zh-CN" mapstructure:"zh-CN" yaml:"zh-CN"`
	EN    string `json:"en"`
}

type Relation []RelationItem

type RelationAction string

func (action RelationAction) Validate() bool {
	if action != RelationActionSHOW && action != RelationActionHIDDEN {
		return false
	}
	return true
}

func (action RelationAction) negation() RelationAction {
	if action == RelationActionSHOW {
		return RelationActionHIDDEN
	}
	if action == RelationActionHIDDEN {
		return RelationActionSHOW
	}
	fmt.Printf("ERROR!, not support argment releation action %s \n", action)
	return "NOT_SUPPORT_" + action
}

type RelationItem struct {
	Action RelationAction
	When   *RelationWhen
}

func (item RelationItem) ValidateDefinition() error {
	var errs Errors
	if !item.Action.Validate() {
		errs = append(errs, NewTemplateDefinitionError(fmt.Sprintf("not support relation action `%s`", item.Action), nil))
	}

	if item.When == nil {
		errs = append(errs, NewTemplateDefinitionError(fmt.Sprintf("when should not be nil"), nil))
	} else {
		err := item.When.ValidateDefinition()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

const (
	RelationActionSHOW   = "show"
	RelationActionHIDDEN = "hidden"
)

type RelationWhen struct {
	Name  string
	Value interface{}
	All   []RelationWhenItem
	Any   []RelationWhenItem
}

func (relation RelationWhen) ValidateDefinition() error {
	flag := 0
	if relation.All != nil {
		flag = flag + 1
	}
	if relation.Any != nil {
		flag = flag + 1
	}
	if relation.Name != "" {
		flag = flag + 1
	}

	if flag == 0 {
		return nil
	}

	if flag > 1 {
		return NewTemplateDefinitionError("not support multi relation when, you can only set `name and value` or  `all` or `any`", nil)
	}

	errs := Errors{}
	if relation.All != nil && len(relation.All) > 0 {
		for i, item := range relation.All {
			if item.Name == "" {
				errs = append(errs, NewTemplateDefinitionError(fmt.Sprintf("relation.all[%d].name should not empty", i), nil))
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

type RelationWhenItem struct {
	Name  string
	Value interface{}
}

func (rel *Relation) IsMathcShowAction(argumentsValues map[string]interface{}) bool {
	if rel == nil || len(*rel) == 0 {
		return true
	}

	relationMap := map[RelationAction]RelationItem{}
	for _, relation := range *rel {
		relationMap[relation.Action] = relation
	}

	var choiceRelation RelationItem
	if len(relationMap) > 1 {
		if _, ok := relationMap[RelationActionSHOW]; !ok {
			fmt.Printf("not found show relation of %#v, we think this is matching show action\n", *rel)
			return true
		}

		choiceRelation = relationMap[RelationActionSHOW]
	} else {
		choiceRelation = (*rel)[0]
	}

	match := choiceRelation.When.match(argumentsValues)
	matchShowAction := true

	if match {
		matchShowAction = choiceRelation.Action == RelationActionSHOW
	} else {
		// 不满足 when 指定的条件时，我们认为 需要执行的action 是 原有action的逻辑非。
		// 故，需要 判断 逻辑非 之后，是否 ==  show, 确定是否满足 show的条件。
		matchShowAction = choiceRelation.Action.negation() == RelationActionSHOW
	}

	return matchShowAction
}

func (when *RelationWhen) match(argumentsValues map[string]interface{}) bool {

	if when == nil {
		return true
	}

	if when.All != nil && len(when.All) > 0 {

		var res = true
		for _, item := range when.All {
			res = res && item.match(argumentsValues)
		}
		return res
	} else if when.Any != nil && len(when.Any) > 0 {

		var res = false
		for _, item := range when.Any {
			res = res || item.match(argumentsValues)
		}
		return res
	} else if when.Name != "" {
		return (&RelationWhenItem{
			Name:  when.Name,
			Value: when.Value,
		}).match(argumentsValues)
	}

	return true
}

func (whenItem *RelationWhenItem) match(argumentsValues map[string]interface{}) bool {
	val, exists := argumentsValues[whenItem.Name]
	if !exists {
		return false
	}

	return fmt.Sprint(val) == fmt.Sprint(whenItem.Value)
}
