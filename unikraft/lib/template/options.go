package template

import "strings"

func WithProjectName(projectName string) TemplateOption {
	return func(t *Template) {
		t.ProjectName = projectName
	}
}

func WithGitInit(gitInit bool) TemplateOption {
	return func(t *Template) {
		t.GitInit = gitInit
	}
}

func WithLibName(libName string) TemplateOption {
	return func(t *Template) {
		t.LibName = libName
	}
}

func WithLibKName(libKName string) TemplateOption {
	return func(t *Template) {
		t.LibKName = strings.ToLower(libKName)
		t.LibKNameUpperCase = strings.ToUpper(libKName)
	}
}

func WithVersion(version string) TemplateOption {
	return func(t *Template) {
		t.Version = version
	}
}

func WithDescription(description string) TemplateOption {
	return func(t *Template) {
		t.Description = description
	}
}

func WithAuthorName(authorName string) TemplateOption {
	return func(t *Template) {
		t.AuthorName = authorName
	}
}

func WithAuthorEmail(authorEmail string) TemplateOption {
	return func(t *Template) {
		t.AuthorEmail = authorEmail
	}
}

func WithProvideMain(provideMain bool) TemplateOption {
	return func(t *Template) {
		t.ProvideMain = provideMain
	}
}

func WithDocs(docs bool) TemplateOption {
	return func(t *Template) {
		t.WithDocs = docs
	}
}

func WithPatchedir(patchedir bool) TemplateOption {
	return func(t *Template) {
		t.WithPatchedir = patchedir
	}
}

func WithInitialBranch(initialBranch string) TemplateOption {
	return func(t *Template) {
		t.InitialBranch = initialBranch
	}
}

func WithCopyrightHolder(copyrightHolder string) TemplateOption {
	return func(t *Template) {
		t.CopyrightHolder = copyrightHolder
	}
}

func WithOriginUrl(origin string) TemplateOption {
	return func(t *Template) {
		t.OriginUrl = origin
	}
}
