package gitlocalinfo

import (
	"context"
	"fmt"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/std/command"
)

var _manifest = manifest.Manifest{
	Category:       "git",
	Name:           "git_local_info",
	Description:    "Get some base infromation about the current git repo",
	Driver:         "go",
	Args:           map[string]string{},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{},
		ReturnValues: []manifest.UsageDesc{},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	m := make(map[string]string)
	// Get remotes
	// upstream	https://github.com/cofxlabs/cofx.git (fetch)
	{
		_args := spec.EntrypointArgs{
			"cmd":            "git remote -v",
			"split":          "",
			"extract_fields": "0,1,2",
			"query_columns":  "c0,c1",
			"query_where":    "c2 like '%fetch%'",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_local_info function", err)
		}
		for _, v := range rets {
			fields := strings.Fields(v)
			if len(fields) == 2 {
				m[fields[0]] = fields[1]
			}
		}
	}

	// Get github org and repo name
	// .e.g https://github.com/skoo87/cofx.git
	origin, ok := m["origin"]
	if ok {
		if strings.Contains(origin, "https://github.com") {
			fields := strings.Split(origin, "/")
			if len(fields) == 5 {
				m["github_org"] = fields[3]
				m["github_repo"] = strings.TrimSuffix(fields[4], ".git")
			}
		}
	}

	// Get current branch
	{
		_args := spec.EntrypointArgs{
			"cmd":            "git branch --show-current",
			"split":          "",
			"extract_fields": "0",
			"query_columns":  "c0",
			"query_where":    "",
		}
		_, ep, _ := command.New()
		rets, err := ep(ctx, bundle, _args)
		if err != nil {
			return nil, fmt.Errorf("%w: in git_local_info function", err)
		}
		for _, v := range rets {
			m["current_branch"] = v
			break
		}
	}
	return m, nil
}
