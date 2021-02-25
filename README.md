## Cloudmanaged operator

### Deploy operators 

This step is responsible for deploy operators:
- cloudmanaged
- postgresql by [Zalando](https://github.com/zalando/postgres-operator)
- mysql by [Presslabs](https://github.com/presslabs/mysql-operator)
- redis by [spotathome.com](https://github.com/spotahome/redis-operator)

1. Clone the repo
    ```shell script
    git clone git@github.com:kuberlogic/operator.git
    cd operator
    ```
2. Create docker registry access token. You can do it here https://docs.quay.io/solution/getting-started.html
3. Make sure you have the minikube installed and running (https://minikube.sigs.k8s.io/docs/start/)
    ```shell script
    minikube status
    ```
4. Make sure you have Go installed (https://golang.org/doc/install).
4. Create a secret in kubernetes to access gitlab registry (learn more here https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#inspecting-the-secret-regcred)
    ```shell script
    docker login quay.io/kuberlogic
    DOCKER_REGISTRY_SERVER=https://quay.io
    DOCKER_USER=<username, the one that was given to you during the onboarding>
    DOCKER_PASSWORD=<access token or password created in step 2>
    kubectl create secret docker-registry quay-registry --docker-server=$DOCKER_REGISTRY_SERVER --docker-username=$DOCKER_USER --docker-password=$DOCKER_PASSWORD
    ```
4. Deploy operators
    ```shell script
   make deploy
    ```
5. For the dev usage access to the apiserver is available via http://localhost:30007

### Create PostgreSQL cluster

```
kubectl create cloudmanaged/src/cm-operator/config/samle/cm-postgresql.yaml
```

### Create MySQL cluster

```
kubectl apply -f https://raw.githubusercontent.com/presslabs/mysql-operator/master/examples/example-cluster-secret.yaml
kubectl create cloudmanaged/src/cm-operator/config/samle/cm-mysql.yaml
```

### Create Redis cluster

```
kubectl create cloudmanaged/src/cm-operator/config/samle/cm-redis.yaml
```

