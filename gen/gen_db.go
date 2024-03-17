package gen

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"path/filepath"

	"github.com/google/subcommands"
)

type GenDBCmd struct {
	dir    string
	idType string
}

func (*GenDBCmd) Name() string     { return "gen_db" }
func (*GenDBCmd) Synopsis() string { return "Generate db" }
func (*GenDBCmd) Usage() string {
	return `gen_db <name>:

  Generate adapter and driver.

`
}

func (p *GenDBCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.dir, "d", "", "directory")
	f.StringVar(&p.idType, "id", "int64", "id type")
}

func (p *GenDBCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	i := struct {
		Name   string
		IDType string
	}{Name: f.Arg(0), IDType: p.idType}

	const entityTemplate = `
package entity

type (
	{{.Name}} struct {
		ID {{.IDType}}
	}
)

func New{{.Name}}(id {{.IDType}}) *{{.Name}} {
	return &{{.Name}}{
		ID : id,
	}
}
`

	const gatewayTemp = `
package gateway

type (
	{{.Name}}DTO struct {
		ID {{.IDType}}
	}
	{{.Name}}Gateway struct {
		repo {{.Name}}AdapterDriverPort
		roRepo RO{{.Name}}AdapterDriverPort
	}
	{{.Name}}AdapterDriverPort interface {
Delete{{.Name}}(ctx context.Context, id {{.IDType}}) (error) 
Create{{.Name}}(ctx context.Context, dto *{{.Name}}DTO) (error) 
	}
	RO{{.Name}}AdapterDriverPort interface {
Fetch{{.Name}}ByID(ctx context.Context, id {{.IDType}}) (*{{.Name}}DTO, error) 
	}
)

func new{{.Name}}DTO(id {{.IDType}}) *{{.Name}}DTO {
	return &{{.Name}}DTO{
		ID: id,
	}
}
		
func New{{.Name}}Gateway(repo {{.Name}}AdapterDriverPort, roRepo RO{{.Name}}AdapterDriverPort) *{{.Name}}Gateway {
	return &{{.Name}}Gateway{
		repo: repo,
		roRepo: roRepo,
	}
}

func (g {{.Name}}Gateway) Create{{.Name}}(ctx context.Context, e *entity.{{.Name}}) (error) {
	dto := new{{.Name}}DTO(e.ID)
	err := g.repo.Create{{.Name}}(ctx, dto)
	if err != nil {
		return err
	}
	e.ID = dto.ID
	return nil
}

func (g {{.Name}}Gateway) Delete{{.Name}}(ctx context.Context, e *entity.{{.Name}}) (error) {
	err := g.repo.Delete{{.Name}}(ctx, e.ID)
	if err != nil {
		return err
	}
	return nil
}

func (g {{.Name}}Gateway) ROFetch{{.Name}}ByID(ctx context.Context, id {{.IDType}}) (*entity.{{.Name}}, error) {
	dto,err := g.roRepo.Fetch{{.Name}}ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	e := entity.New{{.Name}}(dto.ID)
	return e,nil
}
`

	const repoTemp = `
package repo

type (
	{{.Name}}Repo struct {
	}
	RO{{.Name}}Repo struct {
	}
)

func NewRO{{.Name}}Repo() *RO{{.Name}}Repo {
	return &RO{{.Name}}Repo{
	}
}
func New{{.Name}}Repo() *{{.Name}}Repo {
	return &{{.Name}}Repo{
	}
}

func (r {{.Name}}Repo) Create{{.Name}} (ctx context.Context, dto *gateway.{{.Name}}DTO) error {
	panic("implement Create{{.Name}}")
}

func (r {{.Name}}Repo) Delete{{.Name}} (ctx context.Context, id {{.IDType}}) error {
	panic("implement Delete{{.Name}}")
}

func (ro RO{{.Name}}Repo) Fetch{{.Name}}ByID (ctx context.Context, id {{.IDType}} ) (*gateway.{{.Name}}DTO, error) {
	panic("implement Fetch{{.Name}}ByID")
}
`

	var rbuf bytes.Buffer
	var gbuf bytes.Buffer
	var ebuf bytes.Buffer
	if err := template.Must(template.New("entity").Parse(entityTemplate)).Execute(&ebuf, i); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if err := template.Must(template.New("gateway").Parse(gatewayTemp)).Execute(&gbuf, i); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if err := template.Must(template.New("repo").Parse(repoTemp)).Execute(&rbuf, i); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	edir, err := mkdir(".", p.dir, "entity")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := write(filepath.Join(edir, i.Name+".go"), ebuf.Bytes()); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	gdir, err := mkdir(".", p.dir, "gateway")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := write(filepath.Join(gdir, i.Name+"Gateway.go"), gbuf.Bytes()); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	rdir, err := mkdir(".", p.dir, "repo")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := write(filepath.Join(rdir, i.Name+"Repo.go"), rbuf.Bytes()); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
