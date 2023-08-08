package template

import (
	"context"
	"html/template"
	"os"
	"path"
	"strings"
	"time"

	_ "embed"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	//go:embed CODING_STYLE.md.tmpl
	CodingStyleTemplate string

	//go:embed Config.uk.tmpl
	ConfigUkTemplate string

	//go:embed CONTRIBUTING.md.tmpl
	ContributingTemplate string

	//go:embed COPYING.md.tmpl
	CopyingTemplate string

	//go:embed main.c.tmpl
	MainTemplate string

	//go:embed Makefile.uk.tmpl
	MakefileUkTemplate string

	//go:embed README.md.tmpl
	ReadmeTemplate string
)

type Template struct {
	ProjectName       string
	LibName           string
	LibKName          string
	LibKNameUpperCase string
	Version           string
	Description       string
	AuthorName        string
	AuthorEmail       string
	ProvideMain       bool
	WithDocs          bool
	WithPatchedir     bool
	GitInit           bool
	InitialBranch     string
	CopyrightHolder   string
	Year              int
	Commit            string
	OriginUrl         string

	KconfigDependencies []string
	SourceFiles         []string
}

type TemplateOption func(*Template)

func NewTemplate(ctx context.Context, topts ...TemplateOption) (Template, error) {
	var templ Template

	for _, topt := range topts {
		topt(&templ)
	}

	return templ, nil
}

// Generate template using `.tmpl` files and `Template` struct fields.
func (t Template) Generate(ctx context.Context, workdir string) error {
	t.Year = time.Now().Year()
	t.Commit = "Initial commit (blank)"
	if !strings.HasSuffix(workdir, "/") {
		workdir += "/"
	}

	// Parsing all the templates.
	codingStyleTmpl, err := template.New("CondingStyle").Parse(CodingStyleTemplate)
	if err != nil {
		return err
	}

	configUkTmpl, err := template.New("ConfigUk").Parse(ConfigUkTemplate)
	if err != nil {
		return err
	}

	contributingTmpl, err := template.New("Contributing").Parse(ContributingTemplate)
	if err != nil {
		return err
	}

	copyingTmpl, err := template.New("Copying").Parse(CopyingTemplate)
	if err != nil {
		return err
	}

	readmeTmpl, err := template.New("Readme").Parse(ReadmeTemplate)
	if err != nil {
		return err
	}

	makefileUkTmpl, err := template.New("Makefile").Parse(MakefileUkTemplate)
	if err != nil {
		return err
	}

	// Creating projectName directory to store all the template files.
	libDir := path.Join(workdir, t.ProjectName)
	if err = os.Mkdir(libDir, os.ModePerm); err != nil {
		return err
	}

	// Creating template files to store template text.
	codingStyleFile, err := os.Create(path.Join(libDir, "CODING_STYLE.md"))
	if err != nil {
		return err
	}

	configUkFile, err := os.Create(path.Join(libDir, "Config.uk"))
	if err != nil {
		return err
	}

	contributingMdFile, err := os.Create(path.Join(libDir, "CONTRIBUTING.md"))
	if err != nil {
		return err
	}

	copyingMdFile, err := os.Create(path.Join(libDir, "COPYING.md"))
	if err != nil {
		return err
	}

	readmeMdFile, err := os.Create(path.Join(libDir, "README.md"))
	if err != nil {
		return err
	}

	makefileUkfile, err := os.Create(path.Join(libDir, "Makefile.uk"))
	if err != nil {
		return err
	}

	// Executing Templates with Template struct values
	if err = codingStyleTmpl.Execute(codingStyleFile, t); err != nil {
		return err
	}

	if err = configUkTmpl.Execute(configUkFile, t); err != nil {
		return err
	}

	if err = contributingTmpl.Execute(contributingMdFile, t); err != nil {
		return err
	}

	if err = copyingTmpl.Execute(copyingMdFile, t); err != nil {
		return err
	}

	if err = readmeTmpl.Execute(readmeMdFile, t); err != nil {
		return err
	}

	if err = makefileUkTmpl.Execute(makefileUkfile, t); err != nil {
		return err
	}

	if t.ProvideMain {
		mainFile, err := os.Create(path.Join(libDir, "main.c"))
		if err != nil {
			return err
		}

		mainTmpl, err := template.New("Main").Parse(MainTemplate)
		if err != nil {
			return err
		}

		if err = mainTmpl.Execute(mainFile, t); err != nil {
			return err
		}
	}

	if t.WithPatchedir {
		if err = os.Mkdir(path.Join(libDir, "patches"), 0o644); err != nil {
			return err
		}
	}

	if t.GitInit {
		// Save initial commit.
		repo, err := git.PlainInit(libDir, false)
		if err != nil {
			return err
		}

		repoConfig, err := repo.Config()
		if err != nil {
			return err
		}
		repoConfig.Author.Name = t.AuthorName
		repoConfig.Author.Email = t.AuthorEmail
		if err = repo.Storer.SetConfig(repoConfig); err != nil {
			return err
		}

		repoWorktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		_, err = repoWorktree.Add("./")
		if err != nil {
			return err
		}

		_, err = repoWorktree.Commit(t.Commit, &git.CommitOptions{
			All: true,
			Author: &object.Signature{
				Name:  t.AuthorName,
				Email: t.AuthorEmail,
				When:  time.Now(),
			},
			AllowEmptyCommits: true,
		})
		if err != nil {
			return err
		}

		// Creating InitialBranch.
		headRef, err := repo.Head()
		if err != nil {
			return err
		}
		ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+t.InitialBranch), headRef.Hash())
		if err = repo.Storer.SetReference(ref); err != nil {
			return err
		}
		if err = repoWorktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + t.InitialBranch),
		}); err != nil {
			return err
		}

		// Deleting `master` branch.
		ref = plumbing.NewHashReference("refs/heads/master", headRef.Hash())
		if err = repo.Storer.RemoveReference(ref.Name()); err != nil {
			return err
		}
	}

	return nil
}
