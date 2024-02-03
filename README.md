
![Build](https://github.com/Steadfastie/gokube/workflows/Go_build/badge.svg?branch=main)
![gokubelogo](https://github.com/Steadfastie/gokube/assets/68227124/fa1438bf-7a43-466f-b301-f358fb17fd8d)

# gokube
Go and K8s mastery

### Docker-compose
    docker-compose up -d
### Create [KinD](https://kind.sigs.k8s.io/) cluster
    kind create cluster --config=kind-config.yaml
### Deploy to KinD
    kubectl apply -f ./deployment

## Dashboard setup sequience
1. kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
2. kubectl create serviceaccount -n kubernetes-dashboard admin-user
3. kubectl create clusterrolebinding -n kubernetes-dashboard admin-user --clusterrole cluster-admin --serviceaccount=kubernetes-dashboard:admin-user
4. kubectl proxy
5. kubectl -n kubernetes-dashboard create token admin-user
6. copy the token
7. follow the address, paste the token and enjoy!
