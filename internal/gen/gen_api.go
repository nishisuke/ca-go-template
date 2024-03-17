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

type GenAPICmd struct {
	dir string
	fw  string
}

func (*GenAPICmd) Name() string     { return "gen_api" }
func (*GenAPICmd) Synopsis() string { return "Generate api" }
func (*GenAPICmd) Usage() string {
	return `gen_api [-fw <framework name>] <name>:

  Generate controller and usecase.

`
}

func (p *GenAPICmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.dir, "d", "", "directory")
	f.StringVar(&p.fw, "fw", "net/http", "framework. supported values are net/http, echo.")
}

func (p *GenAPICmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	i := struct {
		Name      string
		IsEcho    bool
		IsNetHTTP bool
	}{Name: f.Arg(0), IsEcho: p.fw == "echo", IsNetHTTP: p.fw == "net/http"}

	const usecaseTemp = `
package usecase

type (
	{{.Name}}Usecase struct {
		gateway {{.Name}}UsecaseGatewayPort
	}
	{{.Name}}UsecaseInput struct {
	}
	{{.Name}}UsecaseOutput struct {
	}
	{{.Name}}UsecaseGatewayPort interface {
	}
)

func New{{.Name}}Usecase(gateway {{.Name}}UsecaseGatewayPort) *{{.Name}}Usecase {
	return &{{.Name}}Usecase{
		gateway: gateway,
	}
}

func (u {{.Name}}Usecase) Exec(ctx context.Context, i *{{.Name}}UsecaseInput) (*{{.Name}}UsecaseOutput, error) {
	return &{{.Name}}UsecaseOutput{}, nil
}
`

	const controllerTemp = `
package controller

type (
	{{.Name}}Controller struct {
		usecase *usecase.{{.Name}}Usecase
	}
)

func New{{.Name}}Controller(usecase *usecase.{{.Name}}Usecase) *{{.Name}}Controller {
	return &{{.Name}}Controller{
		usecase: usecase,
	}
}


{{if .IsNetHTTP}}
func (u {{.Name}}Controller) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input usecase.{{.Name}}UsecaseInput
	output, err := u.usecase.Exec(ctx, &input)
	if err != nil {
		// TODO: handle err
	}
	_=output
}
{{end}}

{{if .IsEcho}}
func (u {{.Name}}Controller) EchoHandler(c echo.Context) error {
	ctx := c.Request().Context()
	var input usecase.{{.Name}}UsecaseInput
	output, err := u.usecase.Exec(ctx, &input)
	if err != nil {
		return err
	}
	_=output
	return nil
}
{{end}}
`

	var ubuf bytes.Buffer
	var cbuf bytes.Buffer
	if err := template.Must(template.New("usease").Parse(usecaseTemp)).Execute(&ubuf, i); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}
	if err := template.Must(template.New("controller").Parse(controllerTemp)).Execute(&cbuf, i); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	udir, err := mkdir(".", p.dir, "usecase")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := write(filepath.Join(udir, i.Name+"Usecase.go"), ubuf.Bytes()); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	cdir, err := mkdir(".", p.dir, "controller")
	if err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	if err := write(filepath.Join(cdir, i.Name+"Controller.go"), cbuf.Bytes()); err != nil {
		fmt.Println(err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
