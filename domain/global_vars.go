package domain

import "gitlab.uaus.cn/devops/jenkinsrender/domain/common"

type GlobalVar struct {
	Name        string                `json:"name"`
	Description common.MulitLangValue `json:"description"`
}

var jenkinsVars = []GlobalVar{
	GlobalVar{
		Name: "BUILD_NUMBER",
		Description: common.MulitLangValue{
			ZH_CN: "当前构建的jenkins编号, 例如 153",
			EN:    "The current build number, such as \"153\"",
		},
	},
	GlobalVar{
		Name: "JOB_NAME",
		Description: common.MulitLangValue{
			ZH_CN: "当前流水线的名称",
			EN:    "Name of the project of this build",
		},
	},
	GlobalVar{
		Name: "JOB_URL",
		Description: common.MulitLangValue{
			ZH_CN: "当前流水线所在Jenkins的地址, 例如http://server:port/jenkins/job/foo/ (需要在Jenkins配置Jenkins URL)",
			EN:    "Full URL of this job, like http://server:port/jenkins/job/foo/ (Jenkins URL must be set)",
		},
	},
	GlobalVar{
		Name: "BUILD_URL",
		Description: common.MulitLangValue{
			ZH_CN: "当前构建所在Jenkins的地址, 例如http://server:port/jenkins/job/foo/15 (需要在Jenkins配置Jenkins URL)",
			EN:    "Full URL of this build, like http://server:port/jenkins/job/foo/15 (Jenkins URL must be set)",
		},
	},
}

var gitVars = []GlobalVar{
	GlobalVar{
		Name: "GIT_COMMIT",
		Description: common.MulitLangValue{
			ZH_CN: "代码提交版本号, 例如: c68938922a3500a95b1f33883144196abc5a794d",
			EN:    "GIT commit id of code repository, such as:\"c68938922a3500a95b1f33883144196abc5a794d\"",
		},
	},
	GlobalVar{
		Name: "GIT_BRANCH",
		Description: common.MulitLangValue{
			ZH_CN: "代码提交分支名称",
			EN:    "GIT branch nam of code repository",
		},
	},
}

var repoVars = []GlobalVar{
	GlobalVar{
		Name: "REPOSITORY_PATH",
		Description: common.MulitLangValue{
			ZH_CN: "代码仓库地址",
			EN:    "url of code repository",
		},
	},
}

var svnVars = []GlobalVar{
	GlobalVar{
		Name: "SVN_REVISION",
		Description: common.MulitLangValue{
			ZH_CN: "svn 代码版本号,例如: 46",
			EN:    "code version of svn repository, such as: \"46\"",
		},
	},
}

var imageVars = []GlobalVar{
	GlobalVar{
		Name: "IMAGE_REPOSITORY",
		Description: common.MulitLangValue{
			ZH_CN: "流水线被镜像触发时的镜像名称",
			EN:    "repository of image when pipeline triggered by docker image",
		},
	},
	GlobalVar{
		Name: "IMAGE_TAG",
		Description: common.MulitLangValue{
			ZH_CN: "流水线被镜像触发时的镜像TAG",
			EN:    "tag of image when pipeline triggered by docker image",
		},
	},
}

//GetGlobalVars  get global vars that jenkinsfilext support
func GetGlobalVars(scm *SCMInfo, imageRepositories []string) []GlobalVar {
	var vars = []GlobalVar{}
	vars = append(vars, jenkinsVars...)

	if scm != nil {
		vars = append(vars, repoVars...)
		if scm.Type == SCMTypeEnum.GIT {
			vars = append(vars, gitVars...)
		}
		if scm.Type == SCMTypeEnum.SVN {
			vars = append(vars, svnVars...)
		}
	}

	if len(imageRepositories) > 0 {
		vars = append(vars, imageVars...)
	}

	return vars
}
