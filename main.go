package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"

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
		Load LoadCmd `cmd`
	}
	ctx := kong.Parse(&cli)
	if err := ctx.Run(); err != nil {
		panic(err)
	}
}

// LoadCmd is a command to load models
type LoadCmd struct {
	Path    string   `help:"path to schema package" required:""`
	Models  []string `help:"Models to load"`
	Dialect string   `help:"dialect to use" enum:"mysql,sqlite,postgres" required:""`
	out     io.Writer
}

func (c *LoadCmd) Run(ctx context.Context) error {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule | packages.NeedDeps}
	pkgs, err := packages.Load(cfg, c.Path)
	if err != nil {
		return err
	}
	models := gatherModels(pkgs)
	p := Payload{
		Models:  models,
		Dialect: c.Dialect,
	}
	var buf bytes.Buffer
	if err := loaderTmpl.Execute(&buf, p); err != nil {
		return err
	}
	source, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}
	s, err := runprog(source)
	if err != nil {
		panic(err)
	}
	if c.out == nil {
		c.out = os.Stdout
	}
	_, err = fmt.Fprintln(c.out, s)
	return err
}

func runprog(src []byte) (string, error) {
	if err := os.MkdirAll(".gormschema", os.ModePerm); err != nil {
		return "", err
	}
	target := fmt.Sprintf(".atlasloader/%s.go", filename("hi"))
	if err := os.WriteFile(target, src, 0644); err != nil {
		return "", fmt.Errorf("gormschema: write file %s: %w", target, err)
	}
	defer os.RemoveAll(".atlasloader")
	return gorun(target, nil)
}

// run 'go run' command and return its output.
func gorun(target string, buildFlags []string) (string, error) {
	s, err := gocmd("run", target, buildFlags)
	if err != nil {
		return "", fmt.Errorf("entc/load: %s", err)
	}
	return s, nil
}

// golist checks if 'go list' can be executed on the given target.
func golist(target string, buildFlags []string) error {
	_, err := gocmd("list", target, buildFlags)
	return err
}

// goCmd runs a go command and returns its output.
func gocmd(command, target string, buildFlags []string) (string, error) {
	args := []string{command}
	args = append(args, buildFlags...)
	args = append(args, target)
	cmd := exec.Command("go", args...)
	stderr := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}
	return stdout.String(), nil
}

func filename(pkg string) string {
	name := strings.ReplaceAll(pkg, "/", "_")
	return fmt.Sprintf("atlasloader_%s_%d", name, time.Now().Unix())
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
}

func gatherModels(pkgs []*packages.Package) []model {
	var models []model
	for _, pkg := range pkgs {
		for k, v := range pkg.TypesInfo.Defs {
			_, ok := v.(*types.TypeName)
			if !ok || !k.IsExported() {
				continue
			}
			if isGORMModel(k.Obj.Decl) {
				models = append(models, model{
					ImportPath: pkg.PkgPath,
					Name:       k.Name,
					PkgName:    pkg.Name,
				})
			}
		}
	}
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
	for _, f := range st.Fields.List {
		if len(f.Names) == 0 && embedsModel(f.Type) {
			return true
		}
	}
	// Look for gorm: tag.
	for _, f := range st.Fields.List {
		if f.Tag != nil && reflect.StructTag(f.Tag.Value).Get("gorm") != "" {
			return true
		}
	}
	return false
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
