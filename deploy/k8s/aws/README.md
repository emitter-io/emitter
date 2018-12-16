# Deploying Emitter with Kubernetes on AWS

This directory contains Kubernetes configuration files which can be used to deploy an [production-grade cluster of emitter](https://emitter.io) on Amazon Web Services (EKS).

In order to get emitter running, you'll need to have `kubectl` [tool installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and an [Amazon Web Services](https://aws.amazon.com) account. 

```
kubectl apply -f storage_ssd.yaml
kubectl apply -f service_dns.yaml
kubectl apply -f service_loadbalancer.yaml
kubectl apply -f broker.yaml
```
