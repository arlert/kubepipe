package pipeline

import (
	"io"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	// Scheme ....
	Scheme = runtime.NewScheme()
	// Codecs ...
	Codecs = serializer.NewCodecFactory(Scheme)
	// SchemeGroupVersion ...
	SchemeGroupVersion = schema.GroupVersion{Group: "", Version: "v1"}

	defaultNamespace = "default"
)

func init() {
	addKnownTypes()
}

// Adds the list of known types to api.Scheme.
func addKnownTypes() {
	Scheme.AddKnownTypes(SchemeGroupVersion,
		&v1.Pod{},
		&v1.PersistentVolumeClaim{},
		&v1.Service{},
		&Pipe{},
	)
	return
}

// Parser ....
type Parser struct {
	Pods     map[string]*v1.Pod
	Servcies map[string]*v1.Service
	Pvcs     map[string]*v1.PersistentVolumeClaim
	Pipe     *Pipe
}

// NewParser ...
func NewParser() *Parser {
	return &Parser{
		Pods:     make(map[string]*v1.Pod),
		Servcies: make(map[string]*v1.Service),
		Pvcs:     make(map[string]*v1.PersistentVolumeClaim),
	}
}

//Parse ...
func (p *Parser) Parse(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	decoder := yaml.NewYAMLOrJSONDecoder(file, 4096)
	for {
		objraw := &runtime.RawExtension{}
		err := decoder.Decode(objraw)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		versions := &runtime.VersionedObjects{}

		err = runtime.DecodeInto(Codecs.UniversalDecoder(SchemeGroupVersion), objraw.Raw, versions)
		if err != nil {
			return err
		}
		obj, _ := versions.Last(), versions.First()

		if obj.GetObjectKind().GroupVersionKind().Kind == "Pod" {
			if pod, ok := obj.(*v1.Pod); ok {
				if pod.Namespace == "" {
					pod.Namespace = defaultNamespace
				}
				if pod.Labels == nil {
					pod.Labels = map[string]string{}
				}
				pod.Labels["name"] = pod.Name
				p.Pods[pod.Name] = pod
			}
		} else if obj.GetObjectKind().GroupVersionKind().Kind == "Service" {
			if service, ok := obj.(*v1.Service); ok {
				if service.Namespace == "" {
					service.Namespace = defaultNamespace
				}
				p.Servcies[service.Name] = service
			}
		} else if obj.GetObjectKind().GroupVersionKind().Kind == "PersistentVolumeClaim" {
			if pvc, ok := obj.(*v1.PersistentVolumeClaim); ok {
				if pvc.Namespace == "" {
					pvc.Namespace = defaultNamespace
				}
				p.Pvcs[pvc.Name] = pvc
			}
		} else if obj.GetObjectKind().GroupVersionKind().Kind == "Pipe" {
			if pipe, ok := obj.(*Pipe); ok {
				p.Pipe = pipe
			}
		}
	}
}
