# Deploying Emitter with Kubernetes on DigitalOcean

This directory contains Kubernetes configuration files which can be used to deploy a production-grade cluster on DigitalOcean's Kubernetes.

In order to get emitter running, you'll need to have `kubectl` installed (https://kubernetes.io/docs/tasks/tools/install-kubectl/) and a DigitalOcean account (100$ free credit here: https://m.do.co/c/5bf0e26da5f2). 

```
kubectl --kubeconfig="<your config>" apply -f broker.yaml
kubectl --kubeconfig="<your config>" apply -f service_dns.yaml
kubectl --kubeconfig="<your config>" apply -f service_loadbalancer.yaml
```

The video tutorials below demonstrate how to create an emitter cluster with K8s and DigitalOcean.
[![Creating Kubernetes Cluster](https://img.youtube.com/vi/lgSJCDTubqo/1.jpg)](https://www.youtube.com/watch?v=lgSJCDTubqo)
[![Setting up Emitter Cluster](https://img.youtube.com/vi/CsrKiNjZ2Ew/1.jpg)](https://www.youtube.com/watch?v=CsrKiNjZ2Ew)

