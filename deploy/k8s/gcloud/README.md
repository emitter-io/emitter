# Deploying Emitter with Kubernetes on Google Cloud

This directory contains Kubernetes configuration files which can be used to deploy an [production-grade cluster of emitter](https://emitter.io) on Google Cloud's Kubernetes (GKE).

In order to get emitter running, you'll need to have `kubectl` [tool installed](https://kubernetes.io/docs/tasks/tools/install-kubectl/) and a [Google Cloud](https://cloud.google.com) account. 

```
kubectl apply -f storage_ssd.yaml
kubectl apply -f service_dns.yaml
kubectl apply -f service_loadbalancer.yaml
kubectl apply -f broker.yaml
```


The video tutorials below demonstrate how to create an emitter cluster with K8s and Google Cloud.

[![Creating Kubernetes Cluster](https://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/k8s-gcloud.png)](https://www.youtube.com/watch?v=l1uDWG3Suzw)
[![Setting up Emitter Cluster](http://s3.amazonaws.com/cdn.misakai.com/www-emitter/thumb/emitter-gcloud.png)](https://www.youtube.com/watch?v=IL7WEH_2IOo)

