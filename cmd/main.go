package main

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	"github.com/social-network/subscan/internal/daemons"
	"github.com/social-network/subscan/internal/script"
	"github.com/social-network/subscan/internal/server/http"
	"github.com/social-network/subscan/internal/service"
	"github.com/social-network/substrate-api-rpc/websocket"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func main() {
	defer func() {
		_ = log.Close()
		websocket.CloseWsConnection()
	}()
	if err := setupApp().Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupApp() *cli.App {
	app := cli.NewApp()
	app.Name = "SubScan"
	app.Usage = "SubScan Backend Service, use -h get help"
	app.Version = "1.0"
	app.Action = func(*cli.Context) error { run(); return nil }
	app.Description = "SubScan Backend Service, substrate blockchain explorer"
	app.Flags = []cli.Flag{cli.StringFlag{Name: "conf", Value: "../configs"}}
	app.Before = func(context *cli.Context) error {
		if client, err := paladin.NewFile(context.String("conf")); err != nil {
			panic(err)
		} else {
			paladin.DefaultClient = client
		}
		log.Init(nil)
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	app.Commands = []cli.Command{
		{
			Name: "start",
			Action: func(c *cli.Context) error {
				daemons.Run(c.Args().Get(0), "start")
				return nil
			},
		},
		{
			Name: "stop",
			Action: func(c *cli.Context) error {
				daemons.Run(c.Args().Get(0), "stop")
				return nil
			},
		},
		{
			Name: "install",
			Action: func(c *cli.Context) error {
				script.Install(c.Parent().String("conf"))
				return nil
			},
		},
	}
	return app
}

func run() {
	serviceService := service.New()
	engine := http.New(serviceService)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			if err := engine.Shutdown(ctx); err != nil {
				log.Error("httpSrv.Shutdown error(%v)", err)
			}
			cancel()
			serviceService.Close()
			log.Info("SubScan End exit")
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
