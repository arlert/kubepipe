package pipeline

import (
	"errors"
	"time"

	"github.com/Sirupsen/logrus"
	promise "github.com/fanliao/go-promise"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Pipeline ...
type Pipeline struct {
	*Parser
	*Piperuner
	config *Config
}

// GetClient ...
func GetClient(master, kubeConfigLocation string) (*kubernetes.Clientset, error) {
	// build the config from the master and kubeconfig location
	config, err := clientcmd.BuildConfigFromFlags(master, kubeConfigLocation)
	if err != nil {
		return nil, err
	}

	// creates the clientset
	return kubernetes.NewForConfig(config)
}

// New ...
func New(cfg *Config) (*Pipeline, error) {
	client, err := GetClient("", cfg.KubeConfig)
	if err != nil {
		return nil, err
	}

	return &Pipeline{
		Parser:    NewParser(),
		Piperuner: NewRunner(client),
		config:    cfg,
	}, nil
}

// Run ...
func (p *Pipeline) Run() error {
	// 0. defer clear resource
	defer func() {
		for _, pod := range p.Pods {
			err := p.ClearPod(pod)
			if err != nil {
				logrus.Warnln("clear pod error", err)
			}
		}
		for _, svc := range p.Servcies {
			err := p.ClearService(svc)
			if err != nil {
				logrus.Warnln("clear svc error", err)
			}
		}
		for _, pvc := range p.Pvcs {
			err := p.ClearPvc(pvc)
			if err != nil {
				logrus.Warnln("clear pvc error", err)
			}
		}
	}()

	// Parse file
	err := p.Parse(p.config.Path)
	if err != nil {
		return err
	}
	if p.Pipe == nil {
		return errors.New("pipe not found")
	}
	logrus.Debug("parse config success", p.Pipe)

	// 1. create service /pvcs
	for _, svc := range p.Servcies {
		err := p.CreateService(svc)
		if err != nil {
			return err
		}
	}

	for _, pvc := range p.Pvcs {
		err := p.CreatePvc(pvc)
		if err != nil {
			return err
		}
	}
	// 2. create pod in service
	tasks := []interface{}{}
	for _, name := range p.Pipe.Spec.Services {
		if pod, ok := p.Pods[name]; ok {
			tasks = append(tasks, func() (r interface{}, err error) {
				r, err = p.RunPodUtil(pod, sets.NewString(string(v1.PodRunning)), time.Minute*60, true)
				return
			})
		}
	}
	prepare := promise.WhenAll(tasks...)
	r, err := prepare.Get()
	if err != nil {
		return err
	}
	logrus.Debugln("Prepare complete", r)

	// 3. run pod in stages
	running := promise.Start(func() (r interface{}, err error) {
		logrus.Debug("running start")
		return struct{}{}, nil
	})
	var ok bool
	for _, stage := range p.Pipe.Spec.Stages {
		// now support single job
		if len(stage.Jobs) > 0 {
			logrus.Debugln("stage ", stage)
			if pod, exsit := p.Pods[stage.Jobs[0]]; exsit {
				task := func() (r interface{}, err error) {
					logrus.Debugln("running pod start")
					r, err = p.RunPodUtil(pod,
						sets.NewString(string(v1.PodSucceeded), string(v1.PodFailed)),
						time.Minute*30, true)
					// todo if fail, stop
					logrus.Debugln("running pod error", err)
					return
				}
				_ = task
				running, ok = running.Pipe(task, task)
				if !ok {
					return errors.New("pipe task failed")
				}
			}
		}
	}
	logrus.Debugln("Running start", running)
	r, err = running.Get()
	if err != nil {
		return err
	}
	logrus.Debugln("Running complete")
	// wait for logs
	time.Sleep(time.Second * 5)

	return nil
}
