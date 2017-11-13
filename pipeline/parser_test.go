package pipeline

import (
	"testing"

	"bytes"

	"k8s.io/apimachinery/pkg/util/yaml"
)

func Test_parse(t *testing.T) {
	addKnownTypes()
	decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(`---
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  restartPolicy: Never
  containers:
  - name: pod1
    image: pod1
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  restartPolicy: Never
  containers:
  - name: pod2
    image: pod2
---
apiVersion: v1
kind: Pod
metadata:
  name: pod3
spec:
  restartPolicy: Never
  containers:
  - name: pod3
    image: pod3
---
apiVersion: v1
kind: Pipe
metadata:
  name: pipe
spec:
  env:
  - name: a
    value: a
  - name: b
    value: b
  service:
  - pod3
  stages:
  - name: stage0
    jobs: 
    - pod0
  - name: stage1 
    jobs: pod1
  `)))
	_ = decoder
}
