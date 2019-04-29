package jenkinsfile

import (
	"gitlab.uaus.cn/devops/jenkinsfilext/formatter"
	"bytes"
	"fmt"
	"text/template"

	"github.com/mitchellh/mapstructure"
)

type Pipeline struct {
	Options      *Options
	Agent        interface{}
	Environments []EnvVar
	Stages       []*Stage
	Post         []*PostCondition
}

var (
	POST_ALWAYS   = "always"
	POST_CHANGED  = "changed"
	POST_FAILURE  = "failure"
	POST_SUCCESS  = "success"
	POST_UNSTABLE = "unstable"
	POST_ABORTED  = "aborted"
)

func RenderPipelineAgent(agent interface{}) (string, error) {
	return renderAgent(agent, "agent any")
}

// RenderStageAgent if stage same as pipeline agent , will return empty.
func RenderStageAgent(agent interface{}, pipelineAgent interface{}) (string, error) {
	if EqualAgent(agent, pipelineAgent) {
		return "", nil
	}

	return renderAgent(agent, "")
}

func EqualAgent(agent1 interface{}, agent2 interface{}) bool {
	ag1, err1 := renderAgent(agent1, "")
	ag2, err2 := renderAgent(agent2, "")
	if err1 != nil || err2 != nil {
		return false
	}

	return ag1 == ag2
}

func renderAgent(agent interface{}, defaultAgent string) (string, error) {

	// support string
	if str, ok := agent.(string); ok {
		if str == "" {
			return defaultAgent, nil
		}
		return fmt.Sprintf("agent %s", str), nil
	}
	if agent == nil {
		return defaultAgent, nil
	}

	// support Agent
	var agentStruct = Agent{}
	var ok = false
	if agentStruct, ok = agent.(Agent); !ok {
		err := mapstructure.Decode(agent, &agentStruct)
		if err != nil {
			return "", err
		}
	}

	return agentStruct.Render(defaultAgent), nil
}

type Agent struct {
	Label string
}

func (agent *Agent) Render(defaultAgent string) string {
	if agent == nil {
		return defaultAgent
	}

	if agent.Label != "" {
		return fmt.Sprintf(`agent {label "%s"}`, agent.Label)
	}

	return defaultAgent
}

const pipelineTemplate = `pipeline{

	{{ renderAgent $.Agent}}

	{{- if .Environments}}
	environment{
		{{- range $index, $env := .Environments}}
			{{$env.Name}} = "{{$env.Value}}"
		{{- end}}
	}
	{{- end}}

	options{
		disableConcurrentBuilds()
		buildDiscarder(logRotator(numToKeepStr: '200'))
		{{- if .Options}}
		{{- if .Options.Timeout}}
		timeout(time:{{- .Options.Timeout}}, unit:'SECONDS')
		{{- end}}
		{{- end}}
	}

	stages{
		{{range $i, $stage := .Stages}}
		{{- renderStage $stage}}
		{{- end}}
	}

	post{
		{{range $index, $postCondition := .Post}}
		{{- renderPostCondition $postCondition}}
		{{- end}}
	}
}
`

func (pipeline *Pipeline) initDefault() {
	if pipeline.Agent == nil {
		pipeline.Agent = "any"
	}

	for _, stage := range pipeline.Stages {
		stage.pipeline = pipeline
		for _, pstage := range pipeline.Stages {
			pstage.pipeline = pipeline
		}
	}
}

func (pipeline *Pipeline) Render() (string, error) {
	pipeline.initDefault()

	t, err := template.New("pipeline-template").Funcs(template.FuncMap{
		"renderStage":         RenderStage,
		"join":                Join,
		"renderAgent":         RenderPipelineAgent,
		"renderPostCondition": renderPostCondition,
	}).Parse(pipelineTemplate)

	if err != nil {
		fmt.Printf("parse jenkinsfile pipeline template error:%#v\n", err)
		return "", err
	}

	buffer := bytes.NewBufferString("")
	err = t.Execute(buffer, pipeline)
	if err != nil {
		fmt.Printf("execute jenkinsfile pipeline template error:%#v\n", err)
		return "", err
	}

	return buffer.String(), nil
}

func (pipeline *Pipeline) RenderAndFormat() (render string, err error) {
	render, err = pipeline.Render()
	if err == nil {
		render = formatter.Format(render)
	}
	return
}

type PostCondition struct {
	Name    string
	Scripts string
}

type Stage struct {
	Name         string
	Agent        interface{}
	Options      *Options
	When         *When
	Approve      *Approve
	Environments []EnvVar
	Steps        *Steps

	// 并行时需要
	FailFast bool
	Stages   []*Stage

	pipeline *Pipeline
}

type Options struct {
	Timeout int
}

type Approve struct {
	Timeout int
	Message string
}

type EnvVar struct {
	Name  string
	Value interface{}
}

type When map[string][]string

type Steps struct {
	ScriptsContent string
}

const stageTemplate = `stage("{{- .Name}}"){

	{{ renderAgent $.Agent}}
	{{- if .Environments}}
	environment{
		{{- range $index, $env := .Environments}}
		{{$env.Name}} = "{{$env.Value}}"
		{{- end}}
	}
	{{end}}

	{{- if .When}}
	when{
		beforeAgent true
		{{- range $key, $cdtions := .When}}
			{{- if eq $key "all"}}
		expression { {{join $cdtions " && "}} }
			{{- end}}
			{{- if eq $key "any"}}
		expression { {{join $cdtions "||"}} }
			{{- end}}
		{{- end}}
	}
	{{end}}

	{{- if .Options}}
	options{
		{{- if .Options.Timeout}}
		timeout(time:{{.Options.Timeout}}, unit:'SECONDS')
		{{- end}}
	}
	{{end}}

	{{- $ct := len .Stages}}
	{{- if gt $ct 1  }}
	failFast {{.FailFast}}
	parallel{
		{{- range $i, $stage := .Stages}}
		{{renderStage $stage}}
		{{- end}}
	}
	{{- else}}
	steps{
		{{- if .Approve}}
		timeout(time:{{.Approve.Timeout}}, unit:"SECONDS"){
			input {
				message "{{.Approve.Message}}"
			}
		}
		{{- end}}
		{{ .Steps.ScriptsContent -}}
	}
	{{- end}}
}
`

//RenderStage template func to render stage
func RenderStage(stage Stage) (string, error) {
	return stage.Render()
}

const postConditionTemplate = `{{- .Name}}{
	{{.Scripts}}
}
`

func renderPostCondition(post PostCondition) (string, error) {
	return post.Render()
}

// Join  join string array to string using ch
func Join(arr []string, ch string) string {
	res := ""
	for i, str := range arr {
		if i == 0 {
			res = str
		} else {
			res += ch + str
		}
	}
	return res
}

func (stage *Stage) Render() (string, error) {
	t, err := template.New("stage-template").Funcs(template.FuncMap{
		"renderStage": RenderStage,
		"join":        Join,
		"renderAgent": func(agent interface{}) (string, error) {
			if stage.pipeline != nil {
				return RenderStageAgent(agent, stage.pipeline.Agent)
			}
			return RenderStageAgent(agent, nil)
		},
	}).Parse(stageTemplate)

	if err != nil {
		fmt.Printf("parse jenkinsfile stage template error:%#v\n", err)
		return "", err
	}

	buffer := bytes.NewBufferString("")
	err = t.Execute(buffer, stage)
	if err != nil {
		fmt.Printf("execute jenkinsfile stage template error:%#v\n", err)
		return "", err
	}

	return buffer.String(), nil
}

func (postCondition *PostCondition) Render() (string, error) {
	t, err := template.New("postCondition-template").Parse(postConditionTemplate)

	if err != nil {
		fmt.Printf("parse jenkinsfile post condition template error:%#v\n", err)
		return "", err
	}

	buffer := bytes.NewBufferString("")
	err = t.Execute(buffer, postCondition)
	if err != nil {
		fmt.Printf("execute jenkinsfile post condition template error:%#v\n", err)
		return "", err
	}

	return buffer.String(), nil
}
