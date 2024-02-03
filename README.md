
![Build](https://github.com/steadfastie/gokube/actions/workflows/go.yml/badge.svg?branch=main) ![Release](https://github.com/steadfastie/gokube/actions/workflows/release.yml/badge.svg) ![Publish](https://github.com/steadfastie/gokube/actions/workflows/publish.yml/badge.svg)

![gokubelogo](https://github.com/Steadfastie/gokube/assets/68227124/fa1438bf-7a43-466f-b301-f358fb17fd8d)

# gokube
Go and K8s mastery

### Docker-compose
    docker-compose up -d
### Create [KinD](https://kind.sigs.k8s.io/) cluster
    kind create cluster --config=kind-config.yaml
### Deploy to KinD
    kubectl apply -f ./deployment

### Dashboard setup commands
1.     kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
2.     kubectl apply -f ./deployment/dashboard.yaml
3.     kubectl proxy
4.     kubectl -n kubernetes-dashboard create token admin-user

