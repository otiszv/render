package domain

import (
	"fmt"
	"strings"

	"github.com/otiszv/render/domain/arguments"
	"github.com/otiszv/render/domain/common"
	"github.com/otiszv/render/formatter"

	"github.com/otiszv/render/goutils"
	"github.com/otiszv/render/jenkinsfile"
)

type PipelineTemplateSpec struct {
	Engine       string                `json:"engine"`
	WithSCM      bool                  `json:"withSCM"  mapstructure:"withSCM" yaml:"withSCM"`
	Agent        interface{}           `json:"agent" mapstructure:"agent" yaml:"agent"`
	Stages       []*Stage              `json:"stages"`
	Post         map[string][]*Task    `json:"post"`
	ConstValues  *ConstValues          `json:"values"  mapstructure:"values" yaml:"values"`
	Options      *jenkinsfile.Options  `json:"options"`
	Arguments    arguments.ArgSections `json:"arguments"`
	Environments []jenkinsfile.EnvVar  `json:"environments"`
}

type SCMInfo struct {
	Type           scmType
	RepositoryPath string
	CredentialsID  string
	Branch         string
}

type scmType string

// SCMTypeEnum enum of SCMType
var SCMTypeEnum = struct {
	GIT scmType
	SVN scmType
}{
	GIT: "GIT",
	SVN: "SVN",
}

type ConstValues struct {
	Tasks map[string]*TaskConstValue `json:"tasks"`
}

type TaskConstValue struct {
	Args    map[string]interface{} `json:"args"`
	Options *jenkinsfile.Options   `json:"options"`
	Approve *jenkinsfile.Approve   `json:"approve"`
}

type Stage struct {
	Name       string            `json:"string"`
	Conditions *jenkinsfile.When `json:"conditions"`
	Tasks      []*Task           `json:"tasks"`
}

func (s *Stage) validateDefinition() error {
	if strings.TrimSpace(s.Name) == "" {
		return common.NewTemplateDefinitionError("stage.name should not be empty", nil)
	}

	if s.Tasks == nil || len(s.Tasks) == 0 {
		return common.NewTemplateDefinitionError(fmt.Sprintf("stage `%s`'s tasks should be one at least", s.Name), nil)
	}
	return nil
}

type Task struct {
	Name         string               `json:"name"`
	Agent        interface{}          `json:"agent"`
	Type         string               `json:"type"`
	Options      *jenkinsfile.Options `json:"options"`
	Conditions   *jenkinsfile.When    `json:"conditions"`
	Approve      *jenkinsfile.Approve `json:"approve"`
	Environments []jenkinsfile.EnvVar `json:"environments"`
	Relation     *common.Relation     `json:"relation"`

	taskTemplateSpec      *TaskTemplateSpec
	taskTemplateArgValues map[string]interface{}
	meaningfull           bool
}

func (task *Task) IsMeaningful(argumentsValues map[string]interface{}) bool {
	meaningful := task.Relation.IsMathcShowAction(argumentsValues)
	fmt.Printf("task `%s` meaningful = %t \n", task.Name, meaningful)
	return meaningful
}

func (t *Task) validateDefinition() error {
	errs := common.Errors{}
	if t.Name == "" {
		errs = append(errs, common.NewTemplateDefinitionError("task %s.name shoule not be empty", nil))
	}
	if t.Type == "" {
		errs = append(errs, common.NewTemplateDefinitionError("task %s.type shoule not be empty", nil))
	}

	if err := ValidateAgent(t.Agent); err != nil {
		errs = append(errs, err)
	}

	if strings.Index(t.Name, ".") >= 0 { // name 不能含 .
		errs = append(errs, common.NewTemplateDefinitionError(fmt.Sprintf("task name :%s should not contains dot ", t.Name), nil))
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func unboxToInt(value interface{}) (int, error) {
	v, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("%v is not int", value)
	}
	return v, nil
}

func (t *Task) applyConstValue(constValues *TaskConstValue) {
	if constValues == nil {
		return
	}

	if constValues.Options != nil {
		if t.Options == nil {
			t.Options = &jenkinsfile.Options{}
		}
		t.Options.Timeout = constValues.Options.Timeout
	}

	if constValues.Approve != nil {
		if t.Approve == nil {
			t.Approve = &jenkinsfile.Approve{}
		}
		t.Approve.Timeout = constValues.Approve.Timeout
	}

	if constValues.Args != nil && len(constValues.Args) != 0 {
		for key, value := range constValues.Args {
			if t.taskTemplateArgValues == nil {
				t.taskTemplateArgValues = map[string]interface{}{}
			}
			t.taskTemplateArgValues[key] = value
		}
	}
}

func (t *Task) assignTemplateArgValues(templateArgValues map[string]interface{}) {
	t.taskTemplateArgValues = goutils.MergeMap(templateArgValues, t.taskTemplateArgValues)
}

func (t *Task) assignSystemArgValue(systemArg interface{}) {
	if systemArg != nil {
		t.taskTemplateArgValues = goutils.MergeMap(t.taskTemplateArgValues, map[string]interface{}{
			SystemArgKey: systemArg,
		})
	}
}

func (t *Task) assignArgValues(argValues map[string]interface{}) error {
	for path, value := range argValues {
		err := t.assignArgValueByPath(path, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) assignArgValueByPath(path string, value interface{}) error {
	switch path {
	case "options.timeout":
		v, err := unboxToInt(value)
		if err != nil {
			return common.NewValidateError(fmt.Sprintf("%s's value %v should be int, but got %T", path, value, value), nil)
		}
		t.Options.Timeout = v
	case "approve.timeout":
		v, err := unboxToInt(value)
		if err != nil {
			return err
		}
		t.Options.Timeout = v
	}
	return nil
}

func (t *Task) toJenkinsfileStage() (*jenkinsfile.Stage, error) {
	taskScriptBody, err := t.taskTemplateSpec.Render(t.taskTemplateArgValues)

	if err != nil {
		return nil, err
	}

	var agent = t.Agent
	if agent == nil {
		agent = t.taskTemplateSpec.Agent
	} // if set agent in pipeline template , just use it. if not set in pipeline template but set in task template, use that.

	jenkinsStage := &jenkinsfile.Stage{
		Name:         t.Name,
		Agent:        agent,
		Options:      t.Options,
		When:         t.Conditions,
		Approve:      t.Approve,
		Environments: t.Environments,
		Steps: &jenkinsfile.Steps{
			ScriptsContent: taskScriptBody,
		},
	}

	return jenkinsStage, err
}

// ValidateDefinition validate template define
func (spec *PipelineTemplateSpec) ValidateDefinition() error {
	errs := common.Errors{}

	err := ValidateAgent(spec.Agent)
	if err != nil {
		errs = append(errs, err)
	}

	err = spec.validateStagesDefinition()
	if err != nil {
		errs = append(errs, err)
	}

	err = spec.validateTasksDefinition()
	if err != nil {
		errs = append(errs, err)
	}

	for _, argItem := range spec.Arguments.AllArgItems() {
		err = argItem.ValidateDefinition()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (spec *PipelineTemplateSpec) validateStagesDefinition() error {
	errs := common.Errors{}

	if spec.Stages == nil || len(spec.Stages) == 0 {
		return common.NewTemplateDefinitionError(fmt.Sprint("stages should be one at least"), nil)
	}

	for _, stage := range spec.Stages {
		err := stage.validateDefinition()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (spec *PipelineTemplateSpec) validateTasksDefinition() error {
	errs := common.Errors{}

	nameMap := map[string]interface{}{}

	tasks := spec.allTasks()
	if tasks == nil || len(tasks) == 0 {
		return common.NewTemplateDefinitionError(fmt.Sprint("tasks should be one at least"), nil)
	}

	for _, task := range tasks {
		if _, ok := nameMap[task.Name]; ok {
			errs = append(errs, common.NewTemplateDefinitionError(fmt.Sprintf("task name :%s should be unique", task.Name), nil))
		} else {
			nameMap[task.Name] = struct{}{}
		}
		err := task.validateDefinition()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

//ValidateValue validate values
func (spec *PipelineTemplateSpec) ValidateValue(argumentsValues map[string]interface{}) error {
	argItems := spec.Arguments.AllArgItems()
	argItemsMap := make(map[string]arguments.ArgItem, len(argItems))
	for _, argItem := range argItems {
		argItemsMap[argItem.Name] = argItem
	}

	errs := common.Errors{}
	for argName, value := range argumentsValues {
		if argItem, ok := argItemsMap[argName]; ok {
			if !argItem.IsMeaningful(argumentsValues) {
				fmt.Printf("arg `%s` is not meaningful , skip validate value \n", argItem.Name)
				continue
			}

			err := argItem.ValidateValue(value)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

//Render redner PipelineTemplateSpec to jenkinsfile content
// taskTemplatesRef: you must add `clone` template refs
func (spec *PipelineTemplateSpec) Render(taskTemplatesRef map[string]TaskTemplateSpec, argumentsValues map[string]interface{}, scm *SCMInfo) (string, error) {
	return spec.getRenderEngine()(taskTemplatesRef, argumentsValues, scm)
}

func (spec *PipelineTemplateSpec) RenderAndFormat(taskTemplatesRef map[string]TaskTemplateSpec, argumentsValues map[string]interface{}, scm *SCMInfo) (string, error) {
	pipeline, err := spec.getRenderEngine()(taskTemplatesRef, argumentsValues, scm)

	return formatter.Format(pipeline), err
}

type pipelineTemplateRenderEngine func(taskTemplatesRef map[string]TaskTemplateSpec, argumentsValues map[string]interface{}, scm *SCMInfo) (string, error)

func (spec *PipelineTemplateSpec) getRenderEngine() pipelineTemplateRenderEngine {
	switch spec.Engine {

	default: //default is graph
		{
			return func(taskTemplatesRef map[string]TaskTemplateSpec, argumentsValues map[string]interface{}, scm *SCMInfo) (string, error) {
				return spec.graphRender(taskTemplatesRef, argumentsValues, scm)
			}
		}
	}
}

func (spec *PipelineTemplateSpec) graphRender(taskTemplatesRef map[string]TaskTemplateSpec, argumentsValues map[string]interface{}, scm *SCMInfo) (string, error) {

	err := spec.ValidateDefinition()
	if err != nil {
		return "", err
	}

	// merge default values to argumentsValue
	defaultValues := spec.getDefaultValues()
	argumentsValues = goutils.MergeMap(defaultValues, argumentsValues)

	err = spec.ValidateValue(argumentsValues)
	if err != nil {
		return "", err
	}

	//scm is fixex information
	if spec.WithSCM {
		argumentsValues[CloneTaskTemplateArgName] = scm
		spec.addSCMArg()
	}

	// append task template spec reference
	err = spec.appendTaskTemplateSpecRef(taskTemplatesRef)
	if err != nil {
		return "", err
	}

	// apply const values
	spec.applyConstValues()

	// assign value to all tasks
	spec.assignValuesToEachTask(argumentsValues)

	// mark the task that meaningful
	spec.markMeaningfulTask(argumentsValues)

	pipeline, err := spec.parseToJenkinsfilePipeline()
	if err != nil {
		fmt.Printf("parse to jenkinsfile pipeline error:%#v", err)
		return "", err
	}

	return pipeline.Render()
}

func (spec *PipelineTemplateSpec) addSCMArg() {
	if spec.Arguments == nil || len(spec.Arguments) == 0 {
		spec.Arguments = arguments.ArgSections{
			arguments.ArgSection{
				Items: []arguments.ArgItem{},
			},
		}
	}

	scmArgItem := arguments.ArgItem{
		Name: "SCM",
		Schema: &arguments.ArgItemSchema{
			Type: "object",
		},
		Binding: []string{
			"Clone.args.SCM",
		},
		Required: true,
	}

	spec.Arguments[0].Items = append(spec.Arguments[0].Items, scmArgItem)
}

func (spec *PipelineTemplateSpec) getDefaultValues() (defaultValues map[string]interface{}) {
	defaultValues = map[string]interface{}{}
	allArgItems := spec.Arguments.AllArgItems()
	for _, argItem := range allArgItems {
		if argItem.Default != nil {
			defaultValues[argItem.Name] = argItem.Default
		}
	}

	return
}

func (spec *PipelineTemplateSpec) applyConstValues() {
	if spec.ConstValues == nil {
		return
	}

	if spec.ConstValues.Tasks == nil || len(spec.ConstValues.Tasks) == 0 {
		return
	}

	allTasks := spec.allTasks()
	for _, t := range allTasks {
		t.applyConstValue(spec.ConstValues.Tasks[t.Name])
	}
}

func (spec *PipelineTemplateSpec) parseToJenkinsfilePipeline() (*jenkinsfile.Pipeline, error) {
	jenkinsfileStages, err := spec.getJenkinsfileStages()
	if err != nil {
		fmt.Printf("parse task template script body error :%v\n", err)
		return nil, err
	}
	jenkinsfilePost, err := spec.getJenkinsfilePost()
	if err != nil {
		fmt.Printf("parse task template script body in post error :%v\n", err)
		return nil, err
	}

	return &jenkinsfile.Pipeline{
		Options:      spec.Options,
		Agent:        spec.Agent,
		Environments: spec.Environments,
		Stages:       jenkinsfileStages,
		Post:         jenkinsfilePost,
	}, nil
}

func (spec *PipelineTemplateSpec) getJenkinsfileStages() ([]*jenkinsfile.Stage, error) {
	errs := common.Errors{}

	jenkinsStages := []*jenkinsfile.Stage{}
	for _, stage := range spec.Stages {
		if len(stage.Tasks) == 1 {
			if stage.Tasks[0].meaningfull == false {
				fmt.Printf("task %s is not meaningful, will skip to render it\n", stage.Tasks[0].Name)
				continue
			}

			jenkinsStage, err := stage.Tasks[0].toJenkinsfileStage()
			if err != nil {
				fmt.Printf("render task %s script body error:%#v", stage.Tasks[0].Name, err)
				errs = append(errs, err)
				continue
			}
			jenkinsStages = append(jenkinsStages, jenkinsStage)
		}

		if len(stage.Tasks) > 1 {
			jenkinsStage := &jenkinsfile.Stage{
				Name:   stage.Name,
				When:   stage.Conditions,
				Stages: []*jenkinsfile.Stage{},
			}

			for _, parallelTask := range stage.Tasks {
				if parallelTask.meaningfull == false {
					fmt.Printf("task %s is not meaningful, will skip to render it\n", parallelTask.Name)
					continue
				}

				pStage, err := parallelTask.toJenkinsfileStage()
				if err != nil {
					fmt.Printf("render task %s script body error:%#v", parallelTask.Name, err)
					errs = append(errs, err)
					continue
				}
				jenkinsStage.Stages = append(jenkinsStage.Stages, pStage)
			}

			jenkinsStages = append(jenkinsStages, jenkinsStage)
		}
	}

	if len(errs) > 0 {
		return jenkinsStages, errs
	}

	return jenkinsStages, nil
}

func (spec *PipelineTemplateSpec) getJenkinsfilePost() ([]*jenkinsfile.PostCondition, error) {
	errs := common.Errors{}

	jenkinsPost := []*jenkinsfile.PostCondition{}
	for name, tasks := range spec.Post {
		var scripts string
		for _, task := range tasks {
			taskScriptBody, err := task.taskTemplateSpec.Render(task.taskTemplateArgValues)
			if err != nil {
				fmt.Printf("render task %s script body error:%#v", task.Name, err)
				errs = append(errs, err)
				continue
			}
			scripts += taskScriptBody
		}
		postCondition := &jenkinsfile.PostCondition{
			Name:    name,
			Scripts: scripts,
		}
		jenkinsPost = append(jenkinsPost, postCondition)
	}

	// add cleanWorkspace if post is empty
	if spec.Post == nil {
		postCondition := &jenkinsfile.PostCondition{
			Name:    jenkinsfile.POST_ALWAYS,
			Scripts: CleanWorkspaceScript,
		}
		jenkinsPost = append(jenkinsPost, postCondition)
	}

	if len(errs) > 0 {
		return jenkinsPost, errs
	}

	return jenkinsPost, nil
}

var CleanWorkspaceScript = `
			script{
				echo "clean up workspace"
				deleteDir()
			}
`

type taskValues struct {
	// task  template 中定义的参数的值
	templateArgValues map[string]interface{}
	// pipeline 中引用 task时，的参数的值，例如options.timeout 等
	argValues map[string]interface{}
}

var SystemArgKey = "_system_"

// arguments divide to two fields
// 1. system related: project name, namespace, cluster name. in key=_system_
// 2. not implement yet. _jenkins_ : parellel build, history num.
func getSystemArgumentsValues(argumentsValues map[string]interface{}) interface{} {
	v, ok := argumentsValues[SystemArgKey]
	if !ok {
		return nil
	}
	return v
}

func (spec *PipelineTemplateSpec) assignValuesToEachTask(argumentsValues map[string]interface{}) error {

	var allArgItems = []arguments.ArgItem{}
	for _, argSection := range spec.Arguments {
		allArgItems = append(allArgItems, argSection.Items...)
	}

	var tasksValuesMap = map[string]taskValues{}

	for _, argItem := range allArgItems {
		for _, binding := range argItem.Binding {
			segments := strings.Split(binding, ".")

			if len(segments) <= 1 {
				return common.NewTemplateDefinitionError(fmt.Sprintf("Pipeline template error, %s's Binding format:%s error", argItem.Name, binding), nil)
			}

			var taskName = segments[0]
			if _, ok := tasksValuesMap[taskName]; !ok {
				tasksValuesMap[taskName] = taskValues{
					templateArgValues: map[string]interface{}{},
					argValues:         map[string]interface{}{},
				}
			}

			scope := segments[1]
			switch scope {
			case "args":
				{
					fieldName := segments[2]
					//debug it
					//fmt.Printf("task name :%s , fieldName:%s value:%v\n", taskName, fieldName, argumentsValues[argItem.Name])
					tasksValuesMap[taskName].templateArgValues[fieldName] = argumentsValues[argItem.Name]
				}
			default:
				{
					fieldPath := strings.Join(segments[1:], ".")
					tasksValuesMap[taskName].argValues[fieldPath] = argumentsValues[argItem.Name]
				}
			}
		}
	}

	systemValue := getSystemArgumentsValues(argumentsValues)
	// assignValues
	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			// fmt.Printf("%s templateArgValues is %#v\n", task.Name, tasksValuesMap[task.Name].templateArgValues)
			task.assignTemplateArgValues(tasksValuesMap[task.Name].templateArgValues)
			task.assignSystemArgValue(systemValue)
			task.assignArgValues(tasksValuesMap[task.Name].argValues)
		}
	}

	for _, tasks := range spec.Post {
		for _, task := range tasks {
			// fmt.Printf("%s templateArgValues is %#v\n", task.Name, tasksValuesMap[task.Name].templateArgValues)
			task.assignTemplateArgValues(tasksValuesMap[task.Name].templateArgValues)
			task.assignSystemArgValue(systemValue)
			task.assignArgValues(tasksValuesMap[task.Name].argValues)
		}
	}

	return nil
}

func (spec *PipelineTemplateSpec) findTask(taskName string) *Task {
	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			if task.Name == taskName {
				return task
			}
		}
	}

	return nil
}

func (spec *PipelineTemplateSpec) allTasks() []*Task {
	tasks := []*Task{}

	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (spec *PipelineTemplateSpec) AllTaskTypes() []string {
	var names = []string{}

	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			names = append(names, task.Type)
		}
	}

	return names
}

func (spec *PipelineTemplateSpec) appendTaskTemplateSpecRef(taskTemlateRefs map[string]TaskTemplateSpec) error {

	errs := common.Errors{}
	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			if _, ok := taskTemlateRefs[task.Type]; !ok {
				errs = append(errs, common.NewValidateError(fmt.Sprintf("require definition of task template named:%s", task.Type), nil))
				continue
			}
			taskTemplateSpec := taskTemlateRefs[task.Type]
			task.taskTemplateSpec = &taskTemplateSpec
		}
	}

	for _, tasks := range spec.Post {
		for _, task := range tasks {
			if _, ok := taskTemlateRefs[task.Type]; !ok {
				errs = append(errs, common.NewValidateError(fmt.Sprintf("require definition of task template named:%s", task.Type), nil))
				continue
			}
			taskTemplateSpec := taskTemlateRefs[task.Type]
			task.taskTemplateSpec = &taskTemplateSpec
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func (spec *PipelineTemplateSpec) markMeaningfulTask(argumentsValues map[string]interface{}) {

	for _, stage := range spec.Stages {
		for _, task := range stage.Tasks {
			if task.IsMeaningful(argumentsValues) {
				task.meaningfull = true
			} else {
				task.meaningfull = false
			}
		}
	}

	for _, tasks := range spec.Post {
		for _, task := range tasks {
			if task.IsMeaningful(argumentsValues) {
				task.meaningfull = true
			} else {
				task.meaningfull = false
			}
		}
	}
}
