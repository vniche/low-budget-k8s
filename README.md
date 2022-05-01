# low-budget-k8s

This Pulumi stack provisions a low-budget Kubernetes private cluster on Google Cloud, the estimated monthly minimium cost for this infrastructure is $15 USD at the time of this writing.

## TLDR

Try the one-click deployment, just bear in mind that one will need to [setup a Google Cloud service account](https://www.pulumi.com/registry/packages/gcp/service-account/) and use it authorize the Pulumi Cloud platform to interact with it:

[![Deploy](https://get.pulumi.com/new/button.svg)](https://app.pulumi.com/new?template=https://github.com/vniche/low-budget-gke)

PS.: For environments that need thorough security practices, a tightly configure permissions for the service account is necessary, if but if not, the default compute engine service account every project has by default will do just fine.

## Concepts

Since the whole thing will be based on Kubernetes, if you are not aware of what is it or just interested on reading further on it, the official [Doc](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/) is awsome to go through.

Here are the list of concepts used on this implementation:

- [Developer-first infrastructure with Pulumi](https://redmonk.com/jgovernor/2022/04/27/developer-first-infrastructure-a-new-take-on-infrastucture-as-code/): We will not laverage on all a general-purpose programming language offers but it is there, for example code reusing, if there is a code snippet that are regularly copy and pasted between multiple repositories seems a great opportunity to build a library, not to mention we won't have to go through learning a new syntax or limitations of a DSL;
- [Standard GKE cluster](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-architecture): A general purpose Kubernetes cluster;
- [VPC-native cluster](https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips): A cluster that uses [alias IP address ranges](https://cloud.google.com/vpc/docs/alias-ip)
- [Release channels](https://cloud.google.com/kubernetes-engine/docs/concepts/release-channels): Release channels offer customers the ability to balance between stability and the feature set of the version deployed in the cluster;
- [GKE Dataplane V2](https://cloud.google.com/kubernetes-engine/docs/concepts/dataplane-v2): GKE Dataplane V2 is a dataplane for GKE and Anthos clusters that is optimized for Kubernetes networking.

## Requirements

First things first, let's make sure we are all set up with our Google Cloud account, project and Pulumi itself.

- Google Cloud ([Account](https://console.cloud.google.com/), [Project](https://cloud.google.com/resource-manager/docs/creating-managing-projects#creating_a_project) and [CLI](https://cloud.google.com/sdk/docs/install))
- Pulumi ([Account](http://app.pulumi.com/) and [CLI](https://www.pulumi.com/docs/get-started/install/))

```shell
# to login with google cloud sdk (gcloud)
gcloud auth login

# install kubectl (i recommend this approach instead of other package managers)
gcloud components install kubectl

# to login with pulumi
pulumi login
```

## Hands on

### Connecting Pulumi to GCP

All set? Now we need to authorize Pulumi to interact with Google Cloud, we do so by leveraging on Google Cloud's Application Default Credentials (ADC) locally and also enable some required GCP API on the target GCP project:

```shell
# configures Application Default Credentials on local host
gcloud auth application-default login --project my-gcp-project

# enable apis
gcloud services enable compute.googleapis.com --project=my-gcp-project
gcloud services enable servicenetworking.googleapis.com --project=my-gcp-project
gcloud services enable container.googleapis.com --project=my-gcp-project
```

### Network IP ranges planning

A non-automated task for now, but great for learning on networking. We'll use GCPs own documentation for this, which is pretty complete and straightforward altough extense as it is minimally required: [IP address range planning](https://cloud.google.com/kubernetes-engine/docs/concepts/alias-ips#defaults_limits)

Following the docs and testing custom ranges on this neat [IP Calculator](https://jodies.de/ipcalc), the below was reached:

```text
Nodes IPs range
10.0.0.0/24
HostMin:   10.0.0.1
HostMax:   10.0.0.254
(254 adresses)

Pods IPs range
10.0.4.0/22
HostMin:   10.0.4.1
HostMax:   10.0.7.254
(1022 addresses)

Services IPs range
10.0.8.0/23
HostMin:   10.0.8.1
HostMax:   10.0.9.254
(510 addresses)

Master IPs range
10.0.10.240/28
HostMin:   10.0.10.241
HostMax:   10.0.10.254
(14 addresses)
```

The stack configuration YAML ended up like this for this run:

```yaml
# Pulumi.dev.yaml
config:
  google-native:project: my-gcp-project
  google-native:region: us-central1
  google-native:zone: us-central1-a
  low-budget-k8s:cluster-name: my-first-cluster
  low-budget-k8s:nodes-cidr: 10.0.0.0/24
  low-budget-k8s:pods-cidr: 10.0.4.0/22
  low-budget-k8s:services-cidr: 10.0.8.0/23
  low-budget-k8s:max-pods-per-node: 20

```

### Provisioning

Let's finally

```shell
pulumi up
```

### Enjoy

```shell
# configure local kubeconfig with necessary info to authorize kubectl
gcloud container clusters get-credentials my-first-cluster-123123123 --zone us-central1-a --project my-gcp-project

# list cluster namespaces to ensure everything is fine
kubectl get ns
```
