package create

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/spf13/cobra"

	"kraftkit.sh/cmdfactory"
	"kraftkit.sh/config"
	"kraftkit.sh/packmanager"
	"kraftkit.sh/unikraft/lib/template"
)

type Create struct {
	ProjectName     string `long:"project-name" usage:"Set the project name to the template"`
	LibraryName     string `long:"library-name" usage:"Set the library name to the template"`
	LibraryKName    string `long:"library-kname" usage:"Set the library kname to the template"`
	Version         string `long:"version" short:"v" usage:"Set the library version to the template"`
	Description     string `long:"description"  usage:"Set the description to the template"`
	AuthorName      string `long:"author-name" usage:"Set the author name to the template"`
	AuthorEmail     string `long:"author-email" usage:"Set the author email to the template"`
	InitialBranch   string `long:"initial-branch" usage:"Set the initial branch name to the template"`
	CopyrightHolder string `long:"copyright-holder" usage:"Set the copyright holder name to the template"`
	Origin          string `long:"origin" usage:"Source code origin URL."`
	NoProvideMain   bool   `long:"no-provide-main" usage:"Do not provide provide-main to the template"`
	GitInit         bool   `long:"git-init" usage:"Init git through the creating library"`
	WithPatchedir   bool   `long:"patch-dir" usage:"provide patch directory to the template"`
	SoftPack        bool   `long:"soft-pack" usage:"Softly pack the component so that it is available via kraft list"`
	NoWithDocs      bool   `long:"no-docs" usage:"Do not provide docs to the template"`
	ProjectPath     string `long:"project-path" usage:"Where to create library"`
}

func New() *cobra.Command {
	cmd, err := cmdfactory.New(&Create{}, cobra.Command{
		Short:   "Initialise a library template",
		Use:     "create [FLAGS] [NAME]",
		Aliases: []string{"init"},
		Long: heredoc.Doc(`
		Creates a library template
		`),
		Example: heredoc.Doc(`
			$ kraft lib create
			$ kraft lib create sample-project
			`),
		Annotations: map[string]string{
			cmdfactory.AnnotationHelpGroup: "lib",
		},
	})
	if err != nil {
		panic(err)
	}

	return cmd
}

func (*Create) Pre(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	pm, err := packmanager.NewUmbrellaManager(ctx)
	if err != nil {
		return err
	}

	cmd.SetContext(packmanager.WithPackageManager(ctx, pm))

	return nil
}

func (opts *Create) Run(cmd *cobra.Command, args []string) error {
	var err error

	ctx := cmd.Context()
	if len(args) > 0 {
		opts.ProjectName = args[0]
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !config.G[config.KraftKit](ctx).NoPrompt {
		if !opts.GitInit {
			input := textinput.New("Do you want to intialise library with git [Y/n]:")
			input.Placeholder = "Enter Y for yes"
			inputValue, err := input.RunPrompt()
			if err != nil {
				return err
			}
			if inputValue == "Y" || inputValue == "y" {
				opts.GitInit = true
			}
		}

		if len(opts.ProjectName) == 0 {
			input := textinput.New("Project name:")
			input.Placeholder = "Project name cannot be empty"
			opts.ProjectName, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if len(opts.ProjectPath) == 0 {
			input := textinput.New("Work directory:")
			input.Placeholder = "Where to create library"
			input.InitialValue = cwd
			opts.ProjectPath, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if !opts.SoftPack {
			input := textinput.New("Do you want to package it [Y/n]:")
			input.Placeholder = "Y for yes or n for no"
			inputValue, err := input.RunPrompt()
			if err != nil {
				return err
			}
			if inputValue == "Y" || inputValue == "y" {
				opts.SoftPack = true
			}
		}

		if len(opts.LibraryName) == 0 {
			input := textinput.New("Library name:")
			input.Placeholder = "Library name cannot be empty "
			input.InitialValue = "lib-" + opts.ProjectName
			opts.LibraryName, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if len(opts.LibraryKName) == 0 {
			input := textinput.New("Library kname:")
			input.Placeholder = "Library kname cannot be empty"
			input.InitialValue = "LIB" + strings.ToUpper(strings.ReplaceAll(opts.ProjectName, "-", ""))
			opts.LibraryKName, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		// if len(opts.Description) == 0 {
		// 	input := textinput.New("Description:")
		// 	input.Placeholder = "Description"
		// 	input.InitialValue = ""
		// 	opts.Description, err = input.RunPrompt()
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		if len(opts.Version) == 0 {
			input := textinput.New("Version:")
			input.Placeholder = "Version"
			input.InitialValue = "1.0.0"
			opts.Version, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if len(opts.AuthorName) == 0 {
			input := textinput.New("Author name:")
			input.Placeholder = "Author name cannot be empty"
			input.InitialValue = os.Getenv("USER")
			opts.AuthorName, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if len(opts.AuthorEmail) == 0 {
			input := textinput.New("Author email:")
			input.Placeholder = "Author email cannot be empty"
			cmd := exec.Command("git", "config", "--get", "user.email")
			emailBytes, err := cmd.CombinedOutput()
			if err == nil {
				input.InitialValue = string(emailBytes)
			}
			opts.AuthorEmail, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if opts.GitInit && len(opts.InitialBranch) == 0 {
			input := textinput.New("Initial branch:")
			input.Placeholder = "Initial branch"
			input.InitialValue = "staging"
			opts.InitialBranch, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if opts.GitInit && len(opts.Origin) == 0 {
			input := textinput.New("Origin url (Make sure repository is newly created or empty):")
			input.Placeholder = "Enter origin url"
			opts.Origin, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}

		if len(opts.CopyrightHolder) == 0 {
			input := textinput.New("Copyright holder:")
			input.Placeholder = "Copyright holder cannot be empty"
			input.InitialValue = opts.AuthorName
			opts.CopyrightHolder, err = input.RunPrompt()
			if err != nil {
				return err
			}
		}
	} else {
		var errs []string

		if len(opts.ProjectName) == 0 {
			errs = append(errs, fmt.Errorf("project name cannot be empty").Error())
		}

		if len(opts.Version) == 0 {
			errs = append(errs, fmt.Errorf("version cannot be empty").Error())
		}

		if len(opts.AuthorName) == 0 {
			errs = append(errs, fmt.Errorf("author name cannot be empty").Error())
		}

		if len(opts.AuthorEmail) == 0 {
			errs = append(errs, fmt.Errorf("author email cannot be empty").Error())
		}

		if len(errs) > 0 {
			return fmt.Errorf(strings.Join(errs, "\n"))
		}

		if len(opts.LibraryName) == 0 {
			opts.LibraryName = "lib-" + opts.ProjectName
		}

		if len(opts.LibraryKName) == 0 {
			opts.LibraryKName = "LIB" + strings.ToUpper(strings.ReplaceAll(opts.ProjectName, "-", ""))
		}

		if len(opts.ProjectPath) == 0 {
			opts.ProjectPath = cwd
		}

		if len(opts.CopyrightHolder) == 0 {
			opts.CopyrightHolder = opts.AuthorName
		}
	}

	// Creating instance of Template
	templ, err := template.NewTemplate(ctx,
		template.WithGitInit(opts.GitInit),
		template.WithProjectName(opts.ProjectName),
		template.WithLibName(opts.LibraryName),
		template.WithLibKName(opts.LibraryKName),
		template.WithVersion(opts.Version),
		template.WithDescription(opts.Description),
		template.WithAuthorName(opts.AuthorName),
		template.WithAuthorEmail(opts.AuthorEmail),
		template.WithInitialBranch(opts.InitialBranch),
		template.WithCopyrightHolder(opts.CopyrightHolder),
		template.WithProvideMain(!opts.NoProvideMain),
		template.WithDocs(!opts.NoWithDocs),
		template.WithPatchedir(opts.WithPatchedir),
		template.WithOriginUrl(opts.Origin),
	)
	if err != nil {
		return err
	}

	// Note:
	// If `--git-init` flag is given by user then Git is initialised to the library locally.
	// If `--git-username` and `--git-passowrd` are provided by user then library is pushd to origin.
	if err = templ.Generate(ctx, opts.ProjectPath); err != nil {
		return err
	}

	// Packaging softly.
	if opts.SoftPack {
		packageManager := packmanager.G(ctx)
		if err = packageManager.AddSource(ctx, path.Join(opts.ProjectPath, opts.ProjectName)); err != nil {
			return err
		}
		config.G[config.KraftKit](ctx).Unikraft.Manifests = append(
			config.G[config.KraftKit](ctx).Unikraft.Manifests,
			path.Join(opts.ProjectPath, opts.ProjectName),
		)
		if err := config.M[config.KraftKit](ctx).Write(true); err != nil {
			return err
		}
		if err = packageManager.Update(ctx); err != nil {
			return err
		}
	}

	return nil
}
