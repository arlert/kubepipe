# kubepipe
Exec pipelines on kubernetes

## pipeline.yaml
- 由多个pod spec和一个pipe spec组成
- pod spec是原生的kubernetes pod spec, 而pipe spec用yaml形式描述pod的执行次序。
- stage中的任务并行执行，stage之间串行执行。
- service, env则会在整个执行过程中共享。

一个pipeline.yaml应该是这样的。

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  restartPolicy: Never
  containers:
  - name: c
    image: alpine
    command: ["sh", "-c", "echo running pod1"]
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  restartPolicy: Never
  containers:
  - name: c
    image: alpine
    command: ["sh", "-c", "echo running pod2"]
---
apiVersion: v1
kind: Pod
metadata:
  name: pod3
spec:
  restartPolicy: Never
  containers:
  - name: c
    image: alpine
    command: ["sh", "-c", "echo running pod3"]
---
apiVersion: v1
kind: Pipe
metadata:
  name: pipe
spec:
  stages:
  - name: stage1
    jobs: 
    - pod1
  - name: stage2
    jobs: 
    - pod2
  - name: stage3
    jobs: 
    - pod3

```

## Usage

Set valid KUBECONFIG in env first

```
➜  kubepipe git:(master) ✗ make build_static
➜  kubepipe git:(master) ✗ make/release/kubepipe run -f example/example-1.yaml
running pod1
running pod2
running pod3
```

with debug

```
➜  kubepipe git:(master) ✗ make/release/kubepipe run -f example/example-1.yaml --debug
DEBU[0000] parse config success&{{Pipe v1} {      0 0001-01-01 00:00:00 +0000 UTC <nil> <nil> map[] map[] [] nil [] } {[] [] [{stage1 [pod1]} {stage2 [pod2]} {stage3 [pod3]}]}}
DEBU[0000] Prepare complete []
DEBU[0000] stage  {stage1 [pod1]}
DEBU[0000] stage  {stage2 [pod2]}
DEBU[0000] stage  {stage3 [pod3]}
DEBU[0000] Running start &{3916589616287113937 0xc420406060 0xc420254980}
DEBU[0000] running start
DEBU[0000] running pod start
DEBU[0000] creating pod default pod1
DEBU[0003] watch pod success default pod1
DEBU[0003] pod phase default pod1 Pending
DEBU[0004] pod phase default pod1 Succeeded
DEBU[0004] running pod error <nil>
DEBU[0004] running pod start
DEBU[0004] creating pod default pod2
DEBU[0004] watching pod log success default pod1
running pod1
DEBU[0007] watch pod success default pod2
DEBU[0007] pod phase default pod2 Succeeded
DEBU[0007] running pod error <nil>
DEBU[0007] running pod start
DEBU[0007] watching pod log success default pod2
DEBU[0007] creating pod default pod3
running pod2
DEBU[0010] watch pod success default pod3
DEBU[0010] pod phase default pod3 Succeeded
DEBU[0010] running pod error <nil>
DEBU[0010] Running complete
DEBU[0010] watching pod log success default pod3
running pod3
```