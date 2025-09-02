package main

import (
	_ "embed"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"reflect"
	"slices"
	"strings"
	"text/template"

	"ariga.io/atlas-provider-gorm/gormschema"
	"ariga.io/atlas/sdk/tmplrun"
	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"
)

var (
	//go:embed loader.tmpl
	loader     string
	loaderTmpl = template.Must(template.New("loader").Parse(loader))
)

func main() {
	var cli struct {
		Load LoadCmd `cmd:""`
	}
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err) // nolint: errcheck
		os.Exit(1)
	}
}

// LoadCmd is a command to load models
type LoadCmd struct {
	Path      string   `help:"path to schema package" required:""`
	BuildTags string   `help:"build tags to use" default:""`
	Models    []string `help:"Models to load"`
	Dialect   string   `help:"dialect to use" enum:"mysql,sqlite,postgres,sqlserver,spanner" required:""`
	out       io.Writer
}

var viewDefiner = reflect.TypeOf((*gormschema.ViewDefiner)(nil)).Elem()

func (c *LoadCmd) Run() error {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule | packages.NeedDeps,
	}
	if c.BuildTags != "" {
		cfg.BuildFlags = []string{"-tags=" + c.BuildTags}
	}
	var models []model
	switch pkgs, err := packages.Load(cfg, c.Path, viewDefiner.PkgPath()); {
	case err != nil:
		return fmt.Errorf("loading package: %w", err)
	case len(pkgs) != 2:
		return fmt.Errorf("missing package information for: %s", c.Path)
	default:
		schemaPkg, modelsPkg := pkgs[0], pkgs[1]
		if schemaPkg.PkgPath != viewDefiner.PkgPath() {
			schemaPkg, modelsPkg = pkgs[1], pkgs[0]
		}
		models = gatherModels(modelsPkg, schemaPkg.Types.Scope().
			Lookup(viewDefiner.Name()).Type().
			Underlying().(*types.Interface))
	}
	s, err := tmplrun.New("gormschema", loaderTmpl, tmplrun.WithBuildTags(c.BuildTags)).
		Run(Payload{
			Models:  models,
			Dialect: c.Dialect,
		})
	if err != nil {
		return err
	}
	if c.out == nil {
		c.out = os.Stdout
	}
	_, err = fmt.Fprintln(c.out, s)
	return err
}

type Payload struct {
	Models  []model
	Dialect string
}

func (p Payload) Imports() []string {
	imports := make(map[string]struct{})
	for _, m := range p.Models {
		imports[m.ImportPath] = struct{}{}
	}
	var result []string
	for k := range imports {
		result = append(result, k)
	}
	return result
}

type model struct {
	ImportPath string
	PkgName    string
	Name       string
	Pos        string
}

func (m model) String() string {
	return fmt.Sprintf("%s.%s", m.PkgName, m.Name)
}

func gatherModels(pkg *packages.Package, view *types.Interface) []model {
	var models []model
	for k, v := range pkg.TypesInfo.Defs {
		typ, ok := v.(*types.TypeName)
		if !ok || !k.IsExported() {
			continue
		}
		if isGORMModel(k.Obj.Decl) || types.Implements(typ.Type(), view) {
			p := pkg.Fset.Position(k.Pos())
			models = append(models, model{
				ImportPath: pkg.PkgPath,
				Name:       k.Name,
				PkgName:    pkg.Name,
				Pos:        fmt.Sprintf("%s:%d", p.Filename, p.Line),
			})
		}
	}
	slices.SortFunc(models, func(i, j model) int {
		return strings.Compare(i.Name, j.Name)
	})
	return models
}

func isGORMModel(decl any) bool {
	spec, ok := decl.(*ast.TypeSpec)
	if !ok {
		return false
	}
	st, ok := spec.Type.(*ast.StructType)
	if !ok {
		return false
	}
	return slices.ContainsFunc(st.Fields.List, func(f *ast.Field) bool {
		if len(f.Names) == 0 && embedsModel(f.Type) {
			return true
		}
		// Look for gorm: tag.
		return f.Tag != nil && reflect.StructTag(strings.Trim(f.Tag.Value, "`")).Get("gorm") != ""
	})
}

// return gorm.Model from the selector expression
func embedsModel(ex ast.Expr) bool {
	s, ok := ex.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	if !ok {
		return false
	}
	return id.Name == "gorm" && s.Sel.Name == "Model"
}
