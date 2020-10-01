## Cloudmanaged operator

### Deploy operators 

This step is responsible for deploy operators:
- cloudmanaged
- postgresql by [Zalando](https://github.com/zalando/postgres-operator)
- mysql by [Presslabs](https://github.com/presslabs/mysql-operator)
- redis by [spotathome.com](https://github.com/spotahome/redis-operator)

```
# clone cloudmanaged repo into "cloudmanaged" directory
# cd cloudmanaged/src/cm-operator

# need to create secret for access gitlab registry - see details https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#inspecting-the-secret-regcred
docker login gitlab.corp.cloudlinux.com:5001
# should be generated file $HOME/.docker/config.json 
kubectl create secret generic gitlab-registry --from-file=.dockerconfigjson=$HOME/.docker/config.json --type=kubernetes.io/dockerconfigjson

make deploy
```

### Create PostgreSQL cluster

```
# kubectl create cloudmanaged/src/cm-operator/config/samle/cm-postgresql.yaml
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

