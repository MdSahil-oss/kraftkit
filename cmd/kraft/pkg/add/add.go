package add

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/pack"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft"
	"kraftkit.sh/unikraft/app"
	"kraftkit.sh/unikraft/lib"
)

type Add struct {
	Workdir    string `long:"workdir" short:"w" usage:"workdir location to add pkg to that location"`
	Kraftfile  string `long:"kraftfile" usage:"Set an alternative path of the Kraftfile"`
	FromServer bool   `long:"from-server" usage:"Force package manager to use manifests of the server"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Add{}, cobra.Command{
		Short: "Add unikraft library to the project directory",
		Use:   "add [FLAGS] [PACKAGE|DIR]",
		Args:  cmdfactory.MinimumArgs(1, "package Name is not specified to add to the project"),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "pkg",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *Add) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Add) Run(cmd *cobra.Command, args []string) error {
	var workdir string
	var err error

	if len(args) > 1 {
		workdir = args[1]
	} else {
		workdir = opts.Workdir
	}

	if workdir == "." || workdir == "" {
		workdir, err = os.Getwd()
	}

	if err != nil {
		return err
	}

	nameAndVersion := strings.Split(strings.ToLower(args[0]), ":")
	var version string
	if len(nameAndVersion) > 1 {
		version = nameAndVersion[1]
	}

	ctx := cmd.Context()

	packs, err := packmanager.G(ctx).Catalog(ctx,
		packmanager.WithName(nameAndVersion[0]),
		packmanager.WithTypes(unikraft.ComponentTypeLib),
		packmanager.WithVersion(version),
		packmanager.WithCache(!opts.FromServer),
	)
	if err != nil {
		return err
	}

	if len(packs) != 1 {
		return fmt.Errorf("specified package not found")
	}

	// pulls package
	err = packs[0].Pull(ctx,
		pack.WithPullWorkdir(workdir),
	)
	if err != nil {
		return err
	}

	popts := []app.ProjectOption{}

	if len(opts.Kraftfile) > 0 {
		popts = append(popts, app.WithProjectKraftfile(opts.Kraftfile))
	} else {
		popts = append(popts, app.WithProjectDefaultKraftfiles())
	}

	project, err := app.NewProjectFromOptions(
		ctx,
		append(popts, app.WithProjectWorkdir(workdir))...,
	)
	if err != nil {
		return err
	}

	library, err := lib.NewLibraryFromPackage(ctx, packs[0], version)
	if err != nil {
		return err
	}

	err = project.AddLibrary(ctx, library)
	if err != nil {
		return err
	}

	return project.Save()
}
