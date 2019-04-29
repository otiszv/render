package domain

// some structs definition about kubernetes

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"gitlab.uaus.cn/devops/jenkinsrender/domain/common"

	"github.com/bitly/go-simplejson"
	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
)

type KuberneteKind string

const (
	KuberneteKindPipelineTemplate     KuberneteKind = "PipelineTemplate"
	KuberneteKindPipelineTaskTemplate KuberneteKind = "PipelineTaskTemplate"
)

var SupportKinds = []KuberneteKind{
	KuberneteKindPipelineTemplate, KuberneteKindPipelineTaskTemplate,
}

const (
	APIVersionV1Alpha1 = "devops.windcloud/v1alpha1"
)

var SupportVersions = []string{
	APIVersionV1Alpha1,
}

type Kubernete struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       KuberneteKind          `json:"kind"`
	Metadata   *simplejson.Json       `json:"metadata"`
	Data       map[string]string      `json:"data,omitempy"`
	Spec       *simplejson.Json       `json:"spec"`
	Status     map[string]interface{} `json:"status,omitempy"`
}

func (kube *Kubernete) ValidateName() error {
	var nameRegx = "^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$"
	var validName = regexp.MustCompile(nameRegx)
	if !validName.MatchString(kube.GetName("")) {
		return fmt.Errorf("name should match ^[a-zA-Z]([-a-zA-Z0-9]*[a-zA-Z0-9])?$")
	}
	return nil
}

func (kube *Kubernete) GetName(defaultValue string) string {
	val, err := kube.Metadata.Get("name").String()
	if err != nil {
		fmt.Printf("get name return error from %#v , error:%#v \n", kube, err)
		return defaultValue
	}
	return val
}

func (kube *Kubernete) LoadFromYaml(yamls string) error {
	return yaml.Unmarshal([]byte(yamls), kube)
}

func (kube *Kubernete) LoadFromFile(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}

	byts, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return kube.LoadFromYaml(string(byts))
	}

	return common.Error{
		Message: "only support .yaml or .yml",
	}
}

func (kube *Kubernete) ValidateDefinition() error {

	support := false
	for _, version := range SupportVersions {
		if kube.ApiVersion == version {
			support = true
			break
		}
	}

	if !support {
		return common.NewTemplateDefinitionError(fmt.Sprintf("apiVersion %s is not support now", kube.ApiVersion), nil)
	}

	kindSupport := false
	for _, kind := range SupportKinds {
		if kube.Kind == kind {
			kindSupport = true
		}
	}

	if !kindSupport {
		return common.NewTemplateDefinitionError(fmt.Sprintf("kind %s is not support now", kube.Kind), nil)
	}

	if kube.Spec == nil {
		return common.NewTemplateDefinitionError("spec should be required", nil)
	}

	if kube.Metadata == nil {
		return common.NewTemplateDefinitionError("metadata should not be empty", nil)
	}

	err := kube.ValidateName()
	if err != nil {
		return common.NewTemplateDefinitionError(err.Error(), nil)
	}

	var implementor k8sSpecialResource
	switch kube.Kind {
	case KuberneteKindPipelineTemplate:
		{
			define := JenkinsPipelineTemplateDefinition(*kube)
			implementor = &define
			break
		}
	case KuberneteKindPipelineTaskTemplate:
		{
			define := JenkinsPipelineTaskTemplateDefinition(*kube)
			implementor = &define
			break
		}
	}

	err = implementor.ValidateDefinition()
	if err != nil {
		return err
	}

	return nil
}

type k8sSpecialResource interface {
	ValidateDefinition() error
	AppendStatus(statusLabels ...statusResourceLabel)
}

var UnstableResourceLabel = statusResourceLabel{
	Label: struct {
		Key      string
		InValues []string
	}{
		Key:      "phase",
		InValues: []string{"incubator", "test"},
	},
	Status: struct {
		Key   string
		Value string
	}{
		Key:   "phase",
		Value: "ustable",
	},
}

type statusResourceLabel struct {
	Label struct {
		Key      string
		InValues []string
	}
	Status struct {
		Key   string
		Value string
	}
}

type JenkinsPipelineTemplateDefinition Kubernete

func (definition *JenkinsPipelineTemplateDefinition) ValidateDefinition() error {
	spec, err := definition.PipelineTemplateSpec()
	if err != nil {
		return err
	}
	err = spec.ValidateDefinition()
	if err != nil {
		return err
	}

	metadata, err := definition.PipelineTemplateMetadata()
	if err != nil {
		return common.NewTemplateDefinitionError(fmt.Sprintf("Cannot get meta data from template: %s", err.Error()), nil)
	}
	err = metadata.ValidateDefinition()
	if err != nil {
		return err
	}
	return nil
}

func (definition *JenkinsPipelineTemplateDefinition) PipelineTemplateSpec() (*PipelineTemplateSpec, error) {
	spec := PipelineTemplateSpec{}
	err := mapstructure.Decode(definition.Spec.MustMap(), &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (definition *JenkinsPipelineTemplateDefinition) PipelineTemplateMetadata() (*PipelineTemplateMetadata, error) {
	metadata := PipelineTemplateMetadata{}
	err := mapstructure.Decode(definition.Metadata.MustMap(), &metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (definition *JenkinsPipelineTemplateDefinition) AppendStatus(statusLabels ...statusResourceLabel) {
	metadata, _ := definition.PipelineTemplateMetadata()
	if metadata.Labels == nil {
		metadata.Labels = map[string]interface{}{}
	}
	if definition.Status == nil {
		definition.Status = map[string]interface{}{}
	}

	for _, statusLabel := range statusLabels {
		val := fmt.Sprintf(metadata.GetLabelString(statusLabel.Label.Key, ""))
		for _, v := range statusLabel.Label.InValues {
			if v == val {
				definition.Status[statusLabel.Status.Key] = statusLabel.Status.Value
			}
		}
	}
}

type JenkinsPipelineTaskTemplateDefinition Kubernete

func (definition *JenkinsPipelineTaskTemplateDefinition) ValidateDefinition() error {
	spec, err := definition.PipelineTaskTemplateSpec()
	if err != nil {
		return err
	}
	err = spec.ValidateDefinition()
	if err != nil {
		return err
	}

	metadata, err := definition.PipelineTaskTemplateMetadata()
	if err != nil {
		return err
	}

	err = metadata.ValidateDefinition()
	if err != nil {
		return err
	}
	return nil
}

func (definition *JenkinsPipelineTaskTemplateDefinition) PipelineTaskTemplateSpec() (*TaskTemplateSpec, error) {
	spec := TaskTemplateSpec{}
	err := mapstructure.Decode(definition.Spec.MustMap(), &spec)
	if err != nil {
		return nil, err
	}
	return &spec, nil
}

func (definition *JenkinsPipelineTaskTemplateDefinition) PipelineTaskTemplateMetadata() (*PipelineTemplateMetadata, error) {
	metadata := PipelineTemplateMetadata{}
	err := mapstructure.Decode(definition.Metadata.MustMap(), &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func (definition *JenkinsPipelineTaskTemplateDefinition) AppendStatus(statusLabels ...statusResourceLabel) {
	metadata, _ := definition.PipelineTaskTemplateMetadata()
	if metadata.Labels == nil {
		metadata.Labels = map[string]interface{}{}
	}
	for _, statusLabel := range statusLabels {
		val := fmt.Sprintf(metadata.GetLabelString(statusLabel.Label.Key, ""))
		for _, v := range statusLabel.Label.InValues {
			if v == val {
				metadata.Labels[statusLabel.Status.Key] = statusLabel.Status.Value
			}
		}
	}
}

const (
	AnnotationDisplayNameCN = "windcloud/displayName.zh-CN"
	AnnotationDisplayNameEN = "windcloud/displayName.en"
	AnnotationDescriptionCN = "windcloud/description.zh-CN"
	AnnotationDescriptionEN = "windcloud/description.en"
	AnnotationReadmeCN      = "windcloud/readme.zh-CN"
	AnnotationReadmeEN      = "windcloud/readme.en"
	AnnotationVersion       = "windcloud/version"
	AnnotationStype         = "windcloud/style.icon"
)

type PipelineTemplateMetadata struct {
	Name        string                 `json:"name"`
	Annotations map[string]interface{} `json:"annotations"`
	Labels      map[string]interface{} `json:"labels"`
}

func (metadata *PipelineTemplateMetadata) GetAnnotation(name string) (res string) {
	if metadata.Annotations == nil || len(metadata.Annotations) == 0 {
		return ""
	}
	if val, ok := metadata.Annotations[name].(string); ok {
		res = val
	}
	return res
}

func (metadata *PipelineTemplateMetadata) GetLabelString(name string, defaultV string) (res string) {
	if metadata == nil || len(metadata.Labels) == 0 {
		return defaultV
	}
	if val, ok := metadata.Labels[name].(string); ok {
		res = val
	}
	return res
}

// ValidateDefinition validate PipelineTemplateMetadata definition
func (metadata *PipelineTemplateMetadata) ValidateDefinition() error {
	if strings.TrimSpace(metadata.Name) == "" {
		return common.NewTemplateDefinitionError("metadata.name should be required", nil)
	}

	var requiredAnnotations = []string{
		AnnotationDisplayNameCN,
		AnnotationDisplayNameEN,
		AnnotationVersion,
	}
	errs := common.Errors{}
	for _, name := range requiredAnnotations {
		if strings.TrimSpace(metadata.GetAnnotation(name)) == "" {
			errs = append(errs, common.NewTemplateDefinitionError(fmt.Sprintf("metadata.annotations.[%s] is required", name), nil))
		}
	}

	version := metadata.GetAnnotation(AnnotationVersion)
	if version != "" && !strings.HasPrefix(version, "v") {
		errs = append(errs,
			common.NewTemplateDefinitionError(fmt.Sprintf("metadata.annotations.[%s] should start with \"v\" ", AnnotationVersion), nil),
		)
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}
