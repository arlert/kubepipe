package pipeline

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Config ..
type Config struct {
	Path       string
	KubeConfig string
}

// Pipe ...
type Pipe struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec struct {
		Env []struct {
			Name  string
			Value string
		}
		Services []string
		Stages   []struct {
			Name string
			Jobs []string
		}
	}
}

// DeepCopyObject ...
func (in *Pipe) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil

}

// DeepCopy ...
func (in *Pipe) DeepCopy() *Pipe {
	if in == nil {
		return nil
	}
	out := new(Pipe)
	*out = *in
	return out
}
