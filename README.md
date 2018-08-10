# kuberoll

Kills all pods in a Deployment or ReplicaSet (based on Labels) and waits for replacement pods to start.


```bash
$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
hello-7757bfd49b-s8r28   1/1       Running   0          3m
hello-7757bfd49b-tbqmj   1/1       Running   0          3m
hello-7757bfd49b-vvsdl   1/1       Running   0          3m

$ ./kuberoll --label run=hello
--> Listing pods with label run=hello in namespace "default"
    hello-7757bfd49b-s8r28, hello-7757bfd49b-tbqmj, hello-7757bfd49b-vvsdl
====> Delete hello-7757bfd49b-s8r28....
====> Delete hello-7757bfd49b-tbqmj.......
====> Delete hello-7757bfd49b-vvsdl......

$ kubectl get pods
NAME                     READY     STATUS    RESTARTS   AGE
hello-7757bfd49b-84sfg   1/1       Running   0          1m
hello-7757bfd49b-h48m2   1/1       Running   0          1m
hello-7757bfd49b-h9nlw   1/1       Running   0          1m
```
