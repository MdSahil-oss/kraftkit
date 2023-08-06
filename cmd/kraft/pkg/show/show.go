package show

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/iostreams"
	"kraftkit.sh/manifest"
	"kraftkit.sh/packmanager"
)

type Show struct {
	Output string `long:"output" short:"o" usage:"Set output format" default:"yaml"`
	Cache  bool   `long:"local" short:"l" usage:"Search package details locally" default:"false"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Show{}, cobra.Command{
		Short:   "Shows a Unikraft package",
		Use:     "show [FLAGS] [PACKAGE|DIR]",
		Aliases: []string{"sw"},
		Long: heredoc.Doc(`
			Shows a Unikraft package like library, core etc details
		`),
		Args: cmdfactory.MinimumArgs(1, "package name is not specified to show information"),
		Example: heredoc.Doc(`
			# Shows details for the library nginx
			$ kraft pkg show nginx`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "pkg",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *Show) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Show) Run(cmd *cobra.Command, args []string) error {
	opts.Output = strings.ToLower(opts.Output)
	if len(args) == 0 {
		return fmt.Errorf("package name is not specified to show information")
	} else if opts.Output != "json" && opts.Output != "yaml" {
		return fmt.Errorf("specified output format is not supported")
	}

	ctx := cmd.Context()

	metadata, err := packmanager.G(ctx).Show(ctx, opts.Output, packmanager.WithCache(opts.Cache), packmanager.WithName(args[0]))
	if err != nil {
		return err
	}

	if metadata != nil {
		var byteCode []byte
		var manifestStruct *manifest.Manifest
		var origin string
		value := reflect.ValueOf(metadata).Elem()
		numFields := value.NumField()
		structType := value.Type()

		for i := 0; i < numFields; i++ {
			if structType.Field(i).Name == "Origin" {
				origin = value.Field(i).String()
			}
		}

		if len(origin) > 0 && !strings.HasPrefix(origin, "http") {
			var indexYaml manifest.ManifestIndex
			manifestIndexYamlPath := path.Join(config.G[config.KraftKit](ctx).Paths.Manifests, "index.yaml")
			indexbyteCode, err := os.ReadFile(manifestIndexYamlPath)
			if err != nil {
				return err
			}
			if err = yaml.Unmarshal(indexbyteCode, &indexYaml); err != nil {
				return err
			}
			for _, manifestObj := range indexYaml.Manifests {
				if args[0] == manifestObj.Name {
					manifestYamlPath := path.Join(config.G[config.KraftKit](ctx).Paths.Manifests, manifestObj.Manifest)
					byteCode, err = os.ReadFile(manifestYamlPath)
					manifestStruct, err = manifest.NewManifestFromBytes(ctx, byteCode)
					break
				}
			}
		} else if len(origin) > 0 {
			manifestStruct, err = manifest.NewManifestFromURL(ctx, origin)
			if err != nil {
				return err
			}
			if opts.Output == "json" {
				byteCode, err = json.Marshal(manifestStruct)
			} else {
				byteCode, err = yaml.Marshal(manifestStruct)
			}
		}
		if err != nil {
			return err
		}
		if opts.Output == "json" {
			byteCode, err = json.Marshal(manifestStruct)
		} else {
			byteCode, err = yaml.Marshal(manifestStruct)
		}
		if err != nil {
			return err
		}
		if len(byteCode) > 0 {
			fmt.Fprint(iostreams.G(ctx).Out, string(byteCode)+"\n")
		} else {
			return fmt.Errorf("no manifest found for package %s", args[0])
		}
	}
	return nil
}
