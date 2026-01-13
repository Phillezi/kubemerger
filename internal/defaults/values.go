package defaults

import (
	"slices"
)

const (
	DefaultKubeDir    = "~/.kube"
	DefaultKubeConfig = DefaultKubeDir + "/config"
)

var defaultFilePaths = []string{
	DefaultKubeDir,
}

func DefaultFilePaths() []string {
	return slices.Clone(defaultFilePaths)
}
