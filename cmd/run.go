package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/arlert/kubepipe/pipeline"
)

// Run ...
var Run = cli.Command{
	Name:   "run",
	Usage:  "run with pipeline file",
	Action: run,
	Flags: []cli.Flag{
		cli.BoolFlag{
			EnvVar: "DEBUG",
			Name:   "debug",
			Usage:  "run in debug mode",
		},
		cli.StringFlag{
			EnvVar: "PIPELINE_FILE",
			Name:   "f,file",
			Usage:  "pipeline file",
		},
		cli.StringFlag{
			EnvVar: "KUBECONFIG",
			Name:   "kube-config",
			Usage:  "kube config",
		},
	},
}

func run(c *cli.Context) error {
	// debug level if requested by user
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}
	cfg := &pipeline.Config{}
	cfg.Path = c.String("file")
	cfg.KubeConfig = c.String("kube-config")
	pipe, err := pipeline.New(cfg)
	if err != nil {
		return err
	}

	return pipe.Run()
}
