# Deploying with Kubernetes 

This directory contains various configuration files which help you setup Emitter cluster on Kubernetes. 

In order to get started, you'd first need to register a `fast` storage class into Kubernetes, as the recommended way for Emitter message storage is to use SSDs.

You can simply run these files with ```kubectl apply -f ...```

* [Storage Class for AWS](ssd_aws.yaml)
* [Storage Class for Azure](ssd_azure.yaml)
* [Storage Class for Google Cloud Engine](ssd_gce.yaml)
* [Storage Class for Minikube](ssd_aws.yaml)



Once the storage class is registered,  you can experiment with one of `broker_` configurations. A good starting point is [broker_test.yaml](broker_test.yaml) which can be used to create a simple stateful cluster of 3 nodes on minikube - keep in mind that you'd need to edit the license file as well. 

## Please Contribute

These templates are provided for reference only, contributions are welcome.