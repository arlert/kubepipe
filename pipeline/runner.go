package pipeline

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/Sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// ErrTimeOutRunPod ...
var ErrTimeOutRunPod = errors.New("timeout runing pod")

// Piperuner ...
type Piperuner struct {
	client *kubernetes.Clientset
}

// NewRunner ...
func NewRunner(client *kubernetes.Clientset) *Piperuner {
	return &Piperuner{
		client: client,
	}
}

// RunPodUtil : sync run util phase ->
func (p *Piperuner) RunPodUtil(pod *v1.Pod, phase sets.String,
	timeout time.Duration, printlog bool) (ret *v1.Pod, err error) {
	logrus.Debugln("creating pod", pod.Namespace, pod.Name)
	ret, err = p.client.CoreV1().Pods(pod.Namespace).Create(pod)
	if err != nil {
		return
	}
	stop := make(chan struct{})
	go func() {
		time.Sleep(timeout)
		stop <- struct{}{}
	}()

	// 1. waiting watching
	var podWatcher watch.Interface
	tick := time.Tick(3 * time.Second)
Watch:
	for {
		select {
		case <-stop:
			return nil, ErrTimeOutRunPod
		case <-tick:
			podWatcher, err = p.client.Core().Pods(pod.Namespace).Watch(metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(labels.Set(map[string]string{
					"name": pod.Name,
				})).String(),
				Watch: true,
			})
			if err == nil {
				break Watch
			}
		}
	}
	logrus.Debugln("watch pod success", pod.Namespace, pod.Name)

	// 2. wait and watch for phase
	logopend := false
	var ok bool
	for {
		select {
		case ev := <-podWatcher.ResultChan():
			pod, ok = ev.Object.(*v1.Pod)
			if !ok {
				continue
			}
			logrus.Debugln("pod phase", pod.Namespace, pod.Name, pod.Status.Phase)
			if printlog && !logopend && (pod.Status.Phase != v1.PodPending) {
				logopend = true
				go p.openLog(pod)
			}
			if phase.Has(string(pod.Status.Phase)) {
				return pod, nil
			}
		case <-stop:
			return nil, ErrTimeOutRunPod
		}
	}
}

func (p *Piperuner) openLog(pod *v1.Pod) {
	logrus.Debugln("watching pod log success", pod.Namespace, pod.Name)
	for _, container := range pod.Spec.Containers {
		req := p.client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name,
			&v1.PodLogOptions{Container: container.Name, Follow: true})
		if req == nil {
			logrus.Errorln("GetLogs req nil", pod.Name)
			continue
		}
		reader, err := req.Stream()
		if err != nil {
			logrus.Errorln("req.Stream error", pod.Name, err)
			continue
		} else {
			defer reader.Close()
		}
		if _, err := io.Copy(os.Stdout, reader); err != nil {
			logrus.Warnln("read from reader error", pod.Name, err)
		}
	}
}

// ClearPod : rm/clear
func (p *Piperuner) ClearPod(pod *v1.Pod) (err error) {
	err = p.client.CoreV1().Pods(pod.Namespace).Delete(pod.Name, nil)
	// ingress may not exsit for alb
	if k8serror.IsNotFound(err) {
		logrus.Infoln("pod not found", pod.Name)
		return nil
	}
	return
}

// CreateService ...
func (p *Piperuner) CreateService(svc *v1.Service) (err error) {
	_, err = p.client.CoreV1().Services(svc.Namespace).Create(svc)
	return
}

// ClearService : rm/clear
func (p *Piperuner) ClearService(svc *v1.Service) (err error) {
	err = p.client.CoreV1().Services(svc.Namespace).Delete(svc.Name, nil)
	// ingress may not exsit for alb
	if k8serror.IsNotFound(err) {
		logrus.Infoln("svc not found", svc.Name)
		return nil
	}
	return
}
