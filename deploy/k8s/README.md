# Deploying with Kubernetes 

This directory contains various configuration files which help you setup Emitter cluster on Kubernetes. You can find different setups for different public clous as they leverage different storage classes. Typically, you'd want to use SSD-backed message storage for higher performance, given that Emitter's `ssd` message storage provider is optimised for it.

* [Amazon Web Services (EKS)](aws)
* [Microsoft Azure (AKS)](azure)
* [Google Cloud (GKE)](gcloud)
* [Digital Ocean](digitalocean)

Keep in mind that you'd need to edit the license file in `broker.yaml`. 

## Please Contribute

These templates are provided for reference only, contributions are welcome.