![master](https://github.com/sarun87/k8snetlook/workflows/Build%20&%20Test/badge.svg?branch=master)

# k8snetlook
Kubernetes Network Problem Detector

## Introduction
A simple tool to help debug connetivity issues within a Pod or from a specific host in a live kubernetes cluster easily with a single command. Works with both IPv4 as well as IPv6 K8s stacks.

## Background
When connectivity between two applications within a Kubernetes cluster does not work as expected, it requires specific troubleshooting steps in real time to find the issue. Often times, this involves using network tools such as `ping`, `tcpdump`, `traceroute`, `nslookup` and others to vaildate the plumbing; both within the source Pod as well as on the host that the pod is running.

It gets harder when the application Pod does not have an interactive `shell` to log into or the specific network tools installed as part of the image. [Netshoot](https://github.com/nicolaka/netshoot) helps by providing a docker image with all of the networking tools necessary to _manually_ debug. Most other tools set up test deployments and run network health checks via those deployments rather than using the actual Pods/Service that is exhibiting network issues as source/destination.

k8snetlook aims to automate some of the basic mundane debugging steps in a live k8s cluster. It hopes to help minimize the need to manually intervene and debug the network at the get-go.

## Usage
k8snetlook needs `kubeconfig` to be supplied to it using the `-config` flag or by exporting `KUBECONFIG` enviroment variable. The tool can be used:
1) to run host level checks only, use the `host` subcommand
2) to run pod level checks, use the the `pod` subcommand and the required arguments like PodName with it. See `k8snetlook pod --help` for more information. Pod checks will also run host checks.

Command usage examples

To run host checks alone
```
k8snetlook host -config /etc/kubernetes/admin.yaml
```
To run Pod check which automatically runs host checks as well
```
k8snetlook pod -config /etc/kubernetes/admin.yaml -srcpodname bbox-74d847cb47-xtpdn -srcpodns default -dstpodname nginx-6db489d4b7-9l264 -dstpodns default --externalip 8.8.8.8
```

## Caveats
* Needs to be run as root. This is because raw sockets are needed (`CAP_NET_RAW` privilege) to programmatically implement the `ping` functionality. `udp` socket could be used to remove need for this requirement (TBD?)

* The binary is run on the host where the Pod with connectivity issues are present
* If the tool isn't able to initialize k8s client using specified kubeconfig, the tool will fail (FUTURE? run other tests that don't need k8s information)

## Run as a docker container
Docker image is hosted at sarun87/k8snetlook:<release_tag>. Or build your own docker image using `make docker-image` command

* Command to run the tool as a docker container
```
docker run --privileged --pid=host --net=host -v /var/run/docker.sock:/var/run/docker.sock -v $KUBECONFIG:/kubeconfig.yaml sarun87/k8snetlook:v0.2 /k8snetlook host -config /kubeconfig.yaml
```
Notes:
* The above command assumes that the $KUBECONFIG environment variable is pointing to a valid kubeconfig & mounts it within the container
* Mounts docker socket needed interact with docker daemon.
* Needs privileged context to be set to access pod's network namespace.
* --net=host: Should run in host network namespace, --pid=host: Run in host pid (proc & sys paths are mounted which is needed to obtain handles to Pod's network namespace)

## Download binary & run
64-bit linux binary is available for download from the [Releases](https://github.com/sarun87/k8snetlook/releases/latest) page.

* download binary to a host
```
wget https://github.com/sarun87/k8snetlook/releases/download/v0.2/k8snetlook
```
* Make the downloaded file executable
```
chmod u+x k8snetlook
```
* Run tool using sudo or as root
```
./k8snetlook
'host' or 'pod' subcommand expected

usage: k8snetlook subcommand [sub-command-options] [-config path-to-kube-config]

valid subcommands
  pod       Debug Pod & host networking
  host      Debug host networking only
```

## Run within K8s
There are advantages & disadvantages to running a K8s network debugging tool within k8s. Ease of deployment and not requiring ssh access to the host running the problem pod are clear advantages. But the underlying problem could prevent deployment of k8snetlook on the host (Eg: communication from the host to k8s api server is down).

The [examples folder](https://github.com/sarun87/k8snetlook/tree/master/examples) contains a yaml manifest that creates a K8s `Job` and runs to completion. All of the required RBAC objets and the Job itself is deployed in the `k8snetlook` namespace. Steps to run k8snetlook in k8s:

* Change the value of key `command` under `containers` section with the required arguments to k8snetlook
* Change the value of key `kubernetes.io/hostname` under `nodeSelector` section in the yaml spec to the host on which the problem pod is running. Then run the following commands:
* Apply to cluster using:
```
kubectl apply -f examples/run-k8s.yaml
```
* Check results by pulling logs of the completed job
```
kubectl -n k8snetlook get pods
```
```
kubectl -n k8snetlook logs <pod-name>
```
* Delete k8snetlook after the run
```
kubectl delete -f examples/run-k8s.yaml
```
* To rerun the test, delete k8snetlook and re-apply.

## Checks currently supported by the tool
By having to initialize kubernetes client-set, the tool intrinsically performs API connectivity check via K8s-apiserver's VIP/External Loadbalancer in case of highly available k8s-apiserver clusters. The tool supports pure IPv6 K8s stack as well as IPv4.

| Host Checks                                      | Pod Checks                                              |
| ------------------------------------------------ | ------------------------------------------------------- |
| Default gateway connectivity (icmp)              | Default gateway connectivity (icmp)                     |
| K8s-apiserver ClusterIP check (https)            | K8s-apiserver ClusterIP check (https)                   |
| K8s-apiserver individual endpoints check (https) | K8s-apiserver individual endpoints check (https)        |
| K8s-apiserver health-check api (livez)           | Destination Pod IP connectivity (icmp)                  |
|                                                  | External IP connectivity (icmp)                         |
|                                                  | K8s DNS name lookup check (kubernetes.local)            |
|                                                  | K8s DNS name lookup for specific service check          |
|                                                  | Path MTU discovery between Src & Dst Pod (icmp)         |
|                                                  | Path MTU discovery between Src Pod & External IP (icmp) |
|                                                  | All K8s service endpoints IP connectivity check (icmp)  |

## How to build from source
To build tool from source, run `make` as follows:
```
make all
```
The binary named `k8snetlook` will be built under `root-dir/bin/`

To clean existing binaries and supporting files, run:
```
make clean
```

To speed up development, there is a darwin target defined as well. To build a darwin compatible binary, run:
```
make k8snetlook-osx
```

To create a zipped release binary, run:
```
make release
```

To create a docker image, run:
```
make docker-image
```

## Contribute
PRs welcome :)


