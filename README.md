# Key-Value service

## Create a Docker image:

```shell
docker build -t ameydev/kv-svc:<tag-name>
```

## Deploy the application on Kubernetes:

Following resourses will be created.
- `kg-svc` Deployment
- `kg-svc` Service
- `app-ingress` Ingress

```shell
kubectl apply -f k8s/
```

## Deploy the nginx ingress controller:

In case the nginx ingress controller also needs to be installed.

```shell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.1.0/deploy/static/provider/cloud/deploy.yaml
```

## Access the application:

```shell
APP_URL=`kubectl get ing app-ingress --output="jsonpath={.status.loadBalancer.ingress[0].ip}"`

# Check app metrics
curl $APP_URL/metrics

# get dummy data
curl -v $APP_URL/get/abc-1

```
