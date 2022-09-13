package gobuild

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
	"github.com/cofunclabs/cofunc/pkg/textline"
)

var (
	prefixArg = manifest.UsageDesc{
		Name: "prefix",
		Desc: "By default, the module field contents are read from the 'go.mod' file",
	}
	binFormatArg = manifest.UsageDesc{
		Name: "bin_format",
		Desc: "Specifies the format of the binary file that to be built",
	}
	mainpkgArg = manifest.UsageDesc{
		Name: "find_mainpkg_dirs",
		Desc: `Specifies the dirs to find main package, if there are more than one, separated by ','. If not specified, it will find it from current dir.`,
	}
)

var (
	outcomeRet = manifest.UsageDesc{
		Name: "outcome",
	}
)

var _manifest = manifest.Manifest{
	Name:        "go_build",
	Description: "For building go project that based on 'go mod'",
	Driver:      "go",
	Args: map[string]string{
		binFormatArg.Name: "bin/",
		mainpkgArg.Name:   ".",
	},
	RetryOnFailure: 0,
	Usage: manifest.Usage{
		Args:         []manifest.UsageDesc{prefixArg, binFormatArg, mainpkgArg},
		ReturnValues: []manifest.UsageDesc{outcomeRet},
	},
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	bins, err := parseBinFormats(args.GetStringSlice(binFormatArg.Name))
	if err != nil {
		return nil, err
	}

	module := args.GetString(prefixArg.Name)
	if module == "" {
		var err error
		if module, err = textline.FindFileLine("go.mod", splitSpace, getPrefix); err != nil {
			return nil, err
		}
	}

	mainpkgDirs := args.GetStringSlice(mainpkgArg.Name)
	var mainpkgs []string
	for _, dir := range mainpkgDirs {
		pkgs, err := findMainPkg(module, dir)
		if err != nil {
			return nil, err
		}
		mainpkgs = append(mainpkgs, pkgs...)
	}

	var outcomes []string
	for _, pkg := range mainpkgs {
		for _, bin := range bins {
			dstbin := bin.fullBinPath(filepath.Base(pkg))
			cmd, err := buildCommand(ctx, dstbin, pkg, bundle.Resources.Logwriter)
			if err != nil {
				return nil, err
			}
			cmd.Env = append(os.Environ(), bin.envs()...)
			if err := cmd.Start(); err != nil {
				return nil, err
			}
			if err := cmd.Wait(); err != nil {
				return nil, err
			}
			fmt.Fprintf(bundle.Resources.Logwriter, "---> %s\n", cmd.String())
			outcomes = append(outcomes, dstbin)
		}
	}

	return map[string]string{
		outcomeRet.Name: strings.Join(outcomes, ","),
	}, nil
}

func buildCommand(ctx context.Context, binpath, mainpath string, w io.Writer) (*exec.Cmd, error) {
	var args []string
	args = append(args, "build")
	args = append(args, "-o")
	args = append(args, binpath)
	args = append(args, mainpath)

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = w
	cmd.Stderr = w
	return cmd, nil
}

func splitSpace(c rune) bool {
	return unicode.IsSpace(c)
}

// getPrefix read 'module' field from go.mod file
func getPrefix(fields []string) (string, bool) {
	if len(fields) == 2 && fields[0] == "module" {
		return fields[1], true
	}
	return "", false
}
