# KubeDTN

A CNI plugin for creating digital twin networks in Kubernetes.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

> Prebuilt images can be found on `registry.cn-shanghai.aliyuncs.com/gpx/kubedtn` and `registry.cn-shanghai.aliyuncs.com/gpx/kubedtn-controller`

1. Install Instances of Custom Resources:

```sh
make install
```

2. Build CNI plugin image:

```sh
make cni-docker
```

> For nodes in the cluster to access the docker image, you'll have to manually distribute it using `/hack/distribute-image.sh` or push it to a registry using `make cni-push`. (Default image tag is derived from the latest tag in Git history)

3. Deploy CNI plugin to the cluster:

```sh
make cni-deploy
```

4. Build and push controller image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/kube-dtn:tag
```

5. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/kube-dtn:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Uninstall CNI plugin

To delete the CNI plugin from the cluster:

```sh
make cni-undeploy
```

### Undeploy controller

UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works

The plugin has 4 components:

- CNI plugin executable: `/plugin`
- Daemon: `/daemon`
- Controller: `/controllers`
- CLI: `/cmd`

The plugin is distributed by the daemon, which runs as a DaemonSet in `kubedtn` namespace. The controller runs as a Deployment in `kubedtn-system` namespace. Configuration examples can be found in `/config/samples`.

#### Plugin & Daemon

The plugin and daemon use gRPC and protobuf to communicate with each other. Protobuf definition can be found in `/proto/v1/kube_dtn.proto`. The default port for the daemon is `:51111`.

As described in the protobuf definition, the daemon has the following gRPC APIs:

```protobuf
rpc Get (PodQuery) returns (Pod);
rpc SetAlive (Pod) returns (BoolResponse);
rpc AddLinks (LinksBatchQuery) returns (BoolResponse);
rpc DelLinks (LinksBatchQuery) returns (BoolResponse);
rpc UpdateLinks (LinksBatchQuery) returns (BoolResponse);
rpc SetupPod (SetupPodQuery) returns (BoolResponse);
rpc DestroyPod (PodQuery) returns (BoolResponse);
```

and a Prometheus metrics endpoint at `:51112/metrics`.

#### Controller

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/)
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster

#### CLI

- Build: `make cmd-build`
- Run: `go run cmd/main.go`

### Test the Controller

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

### eBPF TCP/IP Bypass

By default, the plugin will deploy eBPF programs to bypass TCP/IP kernel stack on Linux veths, the code can be found in `/bpf`. To disable this feature, change the environment variable `TCPIP_BYPASS` to `0` in DaemonSet configuration.
