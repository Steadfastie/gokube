
![Build](https://github.com/steadfastie/gokube/actions/workflows/go.yml/badge.svg?branch=main) ![Release](https://github.com/steadfastie/gokube/actions/workflows/release.yml/badge.svg) ![Publish](https://github.com/steadfastie/gokube/actions/workflows/publish.yml/badge.svg)

![gokubelogo](https://github.com/Steadfastie/gokube/assets/68227124/fa1438bf-7a43-466f-b301-f358fb17fd8d)

# :grapes: gokube
This project started to master deployment of Go application to Kubernetes. The code reflects a commitment to emulating production-grade architecture and navigating implementation challenges. With sincere hope, this repository evolves into a useful template. While this page focuses on practical how-tos, the ⚓[wiki](https://github.com/Steadfastie/gokube/wiki) is about whys

Regardless of how one configures their local development environment, running the project may require:

#### Updating swagger docs
Swagger is configured with 🔗[swaggo/swag](https://github.com/swaggo/swag). Docker files are set up to automatically update Swagger docs during the build process. The very same command can be used manually:

    swag init -g api/main.go -o ./api/docs  
    
#### Auth0 credentials for Swagger
🔗[The officail guide](https://auth0.com/docs/quickstart/backend/golang/interactive) does not cover a lot. If you require configured credentials or guidance on configuring them yourself, please feel free to contact me via any of my profile links

 *One is always welcome to open an issue or create a discussion!*

## :watermelon: Docker-compose
Execute the command within the project directory

    docker-compose up -d

## :cherries: Kubernetes
The local developer environment for this project was constructed using:

:link:[KinD](https://kind.sigs.k8s.io/) 

:link:[Strimzi](https://strimzi.io/) 

:link:[MongoDB Community Kubernetes Operator](https://github.com/mongodb/mongodb-kubernetes-operator/tree/master) 


### :open_file_folder: Configure KinD cluster
Execute the command within the project directory

    kind create cluster --config=kind-config.yaml

##### Dashboard setup
1.     kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.7.0/aio/deploy/recommended.yaml
2.     kubectl apply -f ./deployment/dashboard.yaml
3.     kubectl proxy
4.     kubectl -n kubernetes-dashboard create token admin-user --duration=24h

Next, deploy Kafka and MongoDB. When you're ready to deploy gokube services, use the following command

    kubectl apply -f ./deployment

### :open_file_folder: Configure Kafka cluster
The easiest approach would be to follow 🔗[Deploy Strimzi using installation files](https://strimzi.io/quickstarts/) guide. To deploy a cluster use

    kubectl apply -f ./deployment/strimzi-kafka.yaml

An alternative, albeit more intricate, method, allowing manual configuration, is well described 🔗[here](https://strimzi.io/docs/operators/latest/deploying). However it's regrettable that Windows environments is not yet covered

### :open_file_folder: Configure MongoDB cluster
1. Follow 🔗[the official guilde](https://github.com/mongodb/mongodb-kubernetes-operator/blob/master/docs/install-upgrade.md#install-the-operator-using-kubectl) diligently
2. Ensure that the MongoDB user being used within gokube has the `readWriteAnyDatabase` role. Consider configuring it as follows:
```
apiVersion: mongodbcommunity.mongodb.com/v1
kind: MongoDBCommunity
metadata:
  name: example-mongodb
spec:
  members: 3
  type: ReplicaSet
...
  users:
    - name: my-user
      db: admin
...
      roles:
        - name: clusterAdmin
          db: admin
        - name: userAdminAnyDatabase
          db: admin
        - name: readWriteAnyDatabase # keep and eye here
          db: admin
```
