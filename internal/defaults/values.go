package defaults

import (
	"slices"
)

const (
	DefaultKubeDir    = "~/.kube"
	DefaultKubeConfig = DefaultKubeDir + "/config"
)

var defaultIgnorePaths = []string{
	DefaultKubeDir + "/cache",
	DefaultKubeDir + "/kubectx",
}

func DefaultIgnorePaths() []string {
	return slices.Clone(defaultIgnorePaths)
}
