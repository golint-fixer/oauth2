package main

import (
	"runtime"

	"./routers"
	"github.com/codegangsta/cli"
	"github.com/iogo-framework/applications"
	"github.com/iogo-framework/cmd"
	"github.com/iogo-framework/settings"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	cmd := cmd.New()
	cmd.Name = "oauth2"
	cmd.Usage = "Quorums OAuth2 server"
	cmd.Version = "0.0.1"
	cmd.Before = serve
	cmd.Flags = append(cmd.Flags, []cli.Flag{
		cli.StringFlag{"cpu, cpuprofile", "", "cpu profiling", ""},
		cli.StringFlag{"port, p", "8080", "server listening port", ""},
		cli.HelpFlag,
	}...)
	cmd.RunAndExitOnError()
}

func serve(ctx *cli.Context) error {
	var app *applications.Application
	var err error

	settings.Port = ctx.String("port")

	if app, err = applications.New(); err != nil {
		return err
	}

	app.Load(routers.URLs)
	app.Serve()

	return nil
}
