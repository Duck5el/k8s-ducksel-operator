# Ducksel Operator

This repository is created solely for educational and tutorial purposes. The content provided here is intended to serve as a learning resource and does not represent any official, production-ready software.

## Prerequires

- [minikube](https://minikube.sigs.k8s.io/docs/)
- [docker](https://docs.docker.com/)
- [kubectl](https://kubernetes.io/docs/reference/kubectl/)
- [go](https://go.dev/)

## Get Started

Install the dependencys

```shell
go get -v -t -d ./...
```

Update project

```shell
make manifest # Generates the CRD based on the code structure
```

Start the operator

```shell
make run
```

Start you minikube server

```shell
minikube start
```

Create a namespace

```shell
kubectl create namespcae ducksel
```

Apply the CRD

```shell
kubectl apply -f config/crd/bases/api.my.domain_ducksels.yaml 
```

Apply the CR

```shell
kubectl apply -f config/samples/api_v1_ducksel.yaml 
```

Tunnel the minikube

```shell
minikube tunnel
```

Call the nginx website

```shell
xdg-open http://$(kubectl get service ducksel-deployment -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

## Info

This project has been created with:

- [operator-sdk](https://sdk.operatorframework.io/)
- [VSCode](https://code.visualstudio.com/)
- [Ubuntu 22.04](https://ubuntu.com/download/desktop)
