package remove

import (
	"os"

	"github.com/spf13/cobra"
	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft/app"
)

type Remove struct {
	Workdir   string `long:"workdir" short:"w" usage:"workdir location to remove pkg from that location"`
	Kraftfile string `long:"kraftfile" usage:"Set an alternative path of the Kraftfile"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Remove{}, cobra.Command{
		Short:   "Remove unikraft library from the project directory",
		Use:     "remove [FLAGS] [PACKAGE] [DIR]",
		Aliases: []string{"rm"},
		Args:    cmdfactory.MinimumArgs(1, "package name is not specified to remove from the project"),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "pkg",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (opts *Remove) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Remove) Run(cmd *cobra.Command, args []string) error {
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
	ctx := cmd.Context()

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

	if err = project.RemoveLibrary(ctx, args[0]); err != nil {
		return err
	}

	return project.Save()
}
