package daemon

import (
	"context"
	"log"
	"os"

	"github.com/phillezi/kubemerger/internal/defaults"
	"github.com/phillezi/kubemerger/pkg/merge"
	"github.com/phillezi/kubemerger/pkg/recwatch"
	"github.com/phillezi/kubemerger/utils"
	"k8s.io/client-go/tools/clientcmd"
)

type Daemon struct {
	ctx  context.Context
	root string
	out  string
}

type Option = func(*Daemon)

func WithContext(ctx context.Context) Option {
	return func(d *Daemon) {
		d.ctx = ctx
	}
}

func WithRoot(rootDir string) Option {
	return func(d *Daemon) {
		d.root = rootDir
	}
}

func WithOutput(output string) Option {
	return func(d *Daemon) {
		d.out = output
	}
}

func New(opts ...Option) *Daemon {
	d := &Daemon{}

	for _, opt := range opts {
		opt(d)
	}

	if d.ctx == nil {
		d.ctx = context.Background()
	}

	if d.root == "" {
		d.root = defaults.DefaultKubeDir
	}

	if d.out == "" {
		d.out = defaults.DefaultKubeConfig
	}

	return d
}

func (d *Daemon) Run() error {
	var output string = utils.Expand(d.out)
	w, err := recwatch.New(
		utils.Expand(d.root),
		append(defaults.DefaultIgnorePaths(), output),
	)
	if err != nil {
		log.Default().Println("Error:", err)
		return err
	}

	for {
		select {
		case files := <-w.FilesCh:
			log.Default().Println("updated files:", files)
			merged, err := merge.MergeFiles(files...)
			if err != nil {
				log.Default().Println("Error merging:", err)
				break
			}
			if merged == nil {
				log.Default().Println("Merged config is nil")
				break
			}
			data, err := clientcmd.Write(*merged)
			if err != nil {
				log.Default().Println("Error marshalling:", err)
				break
			}
			if err := os.WriteFile(output, data, 0o644); err != nil {
				log.Default().Println("Error writing to file:", output, ", err:", err)
			} else {
				log.Default().Println("Wrote to file:", output)
			}
		case <-d.ctx.Done():
			return d.ctx.Err()
		}
	}
}
