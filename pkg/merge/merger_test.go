package merge_test

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phillezi/kubemerger/pkg/merge"
	"k8s.io/client-go/tools/clientcmd/api"
)

const kubeconfigA = `
apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: https://a.example.com
users:
- name: user
  user:
    token: token-a
contexts:
- name: ctx
  context:
    cluster: cluster
    user: user
current-context: ctx
`

const kubeconfigB = `
apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: https://b.example.com
users:
- name: user
  user:
    token: token-b
contexts:
- name: ctx
  context:
    cluster: cluster
    user: user
`

func TestMergeReaders_BasicMerge(t *testing.T) {
	cfg, err := merge.MergeReaders(
		merge.NamedReader{Name: "a", Reader: strings.NewReader(kubeconfigA)},
		merge.NamedReader{Name: "b", Reader: strings.NewReader(kubeconfigB)},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHasKey(t, cfg.Clusters, "a-cluster")
	assertHasKey(t, cfg.Clusters, "b-cluster")

	assertHasKey(t, cfg.AuthInfos, "a-user")
	assertHasKey(t, cfg.AuthInfos, "b-user")

	assertHasKey(t, cfg.Contexts, "a-ctx")
	assertHasKey(t, cfg.Contexts, "b-ctx")
}

func TestMergeReaders_CurrentContext_FirstWins(t *testing.T) {
	cfg, err := merge.MergeReaders(
		merge.NamedReader{Name: "a", Reader: strings.NewReader(kubeconfigA)},
		merge.NamedReader{Name: "b", Reader: strings.NewReader(kubeconfigB)},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.CurrentContext != "a-ctx" {
		t.Fatalf("expected current-context to be a-ctx, got %q", cfg.CurrentContext)
	}
}

func TestMergeReaders_NameCollision(t *testing.T) {
	cfg, err := merge.MergeReaders(
		merge.NamedReader{Name: "same", Reader: strings.NewReader(kubeconfigA)},
		merge.NamedReader{Name: "same", Reader: strings.NewReader(kubeconfigA)},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertHasKey(t, cfg.Clusters, "same-cluster")
	assertHasKey(t, cfg.Clusters, "same-cluster-2")

	assertHasKey(t, cfg.AuthInfos, "same-user")
	assertHasKey(t, cfg.AuthInfos, "same-user-2")

	assertHasKey(t, cfg.Contexts, "same-ctx")
	assertHasKey(t, cfg.Contexts, "same-ctx-2")
}

func TestEnsureTrailingDash(t *testing.T) {
	tests := map[string]string{
		"a":   "a-",
		"a-":  "a-",
		"":    "-",
		"foo": "foo-",
	}

	for in, want := range tests {
		if got := ensureTrailingDash(in); got != want {
			t.Fatalf("ensureTrailingDash(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFilePrefix(t *testing.T) {
	tests := map[string]string{
		"/tmp/config.yaml": "config",
		"config":           "config",
		"a.b.c.yaml":       "a.b.c",
	}

	for in, want := range tests {
		if got := filePrefix(in); got != want {
			t.Fatalf("filePrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestUniqueName(t *testing.T) {
	m := map[string]*api.Cluster{
		"a":   {},
		"a-2": {},
	}

	got := uniqueName(m, "a")
	if got != "a-3" {
		t.Fatalf("uniqueName returned %q, want a-3", got)
	}
}

func assertHasKey[T any](t *testing.T, m map[string]T, key string) {
	t.Helper()
	if _, ok := m[key]; !ok {
		t.Fatalf("expected key %q to exist, \n%v", key, m)
	}
}

func ensureTrailingDash(s string) string {
	if strings.HasSuffix(s, "-") {
		return s
	}
	return s + "-"
}

func filePrefix(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
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
