package merge

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NamedReader associates a kubeconfig reader with a stable name
// used for prefixing merged objects.
type NamedReader struct {
	Name   string
	Reader io.Reader
}

// MergeReaders loads and merges multiple kubeconfigs from readers.
// Names are prefixed by NamedReader.Name and made globally unique.
func MergeReaders(readers ...NamedReader) (*api.Config, error) {
	merged := api.NewConfig()

	for _, r := range readers {
		if err := mergeFromReader(merged, r); err != nil {
			return nil, fmt.Errorf("merge %s: %w", r.Name, err)
		}
	}

	return merged, nil
}

// MergeFiles is a convenience wrapper around MergeReaders that
// accepts file paths.
func MergeFiles(paths ...string) (*api.Config, error) {
	readers := make([]NamedReader, 0, len(paths))

	for _, path := range paths {
		f, err := os.OpenFile(path, os.O_RDONLY, 0o644)
		if err != nil {
			log.Default().Printf("err: %v", err)
			return nil, err
		}

		readers = append(readers, NamedReader{
			Name:   filePrefix(path),
			Reader: f,
		})
	}

	return MergeReaders(readers...)
}

func mergeFromReader(dest *api.Config, r NamedReader) error {
	data, err := io.ReadAll(r.Reader)
	if err != nil {
		return err
	}

	cfg, err := clientcmd.Load(data)
	if err != nil {
		return err
	}

	prefix := r.Name

	clusterMap := make(map[string]string)
	userMap := make(map[string]string)

	// clusters
	for name, cluster := range cfg.Clusters {
		newName := uniqueName(dest.Clusters, notDefault(name, prefix))
		dest.Clusters[newName] = cluster
		clusterMap[name] = newName
	}

	// users
	for name, user := range cfg.AuthInfos {
		newName := uniqueName(dest.AuthInfos, notDefault(name, prefix))
		dest.AuthInfos[newName] = user
		userMap[name] = newName
	}

	// contexts
	for name, ctx := range cfg.Contexts {
		newName := uniqueName(dest.Contexts, notDefault(name, prefix))
		dest.Contexts[newName] = &api.Context{
			Cluster:   clusterMap[ctx.Cluster],
			AuthInfo:  userMap[ctx.AuthInfo],
			Namespace: ctx.Namespace,
		}

		if dest.CurrentContext == "" && cfg.CurrentContext == name {
			dest.CurrentContext = newName
		}
	}

	if dest.APIVersion == "" {
		dest.APIVersion = "v1"
	}
	if dest.Kind == "" {
		dest.Kind = "Config"
	}

	return nil
}

func notDefault(name, prefix string) string {
	if name != "default" {
		name = strings.Join([]string{prefix, name}, "-")
	} else {
		name = prefix
	}
	return name
}

func filePrefix(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func ensureTrailingDash(s string) string {
	if strings.HasSuffix(s, "-") {
		return s
	}
	return s + "-"
}

// uniqueName ensures the key does not already exist in the target map.
// If it does, it appends -2, -3, etc.
func uniqueName[T any](m map[string]T, base string) string {
	if _, ok := m[base]; !ok {
		return base
	}

	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if _, ok := m[candidate]; !ok {
			return candidate
		}
	}
}
