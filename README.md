# dispatch-demo
This is a sample repository which demonstrates, compositionally, how to orchestrate OpenFGA Check resolution using a dispatching pattern where Check subproblems are dispatched across a cluster of peers which implement a common DispatchService RPC definition.

I've simulated local graph resolution with hard coded checks on `document:1#viewer` which is rewritten to `document:1#editor` and `document:1#editor` is hard coded to return `{allowed: true}`. This is just to simulate graph resolution with rewrite values, but in practice the local dispatcher would be our current equivalent of the graph resolution algorithm for Check.

Here's how it works:
1. Each FGA server implements two RPC server definitions: 1) a public facing gRPC [FGAService](./proto/fga/v1/fga_service.proto) and 2) an internally facing gRPC [DispatchService](./proto/dispatch/v1/dispatch_service.proto).

   The FGAService definition is the API which external clients use. Internally we resolve FGAService/Check requests using internal peer to peer dispatching by delegating subproblems to the DispatchService/DispatchCheck RPC which each peer implements.

   We use consistent hashing to dispatch Check subproblems to peers that own that subproblem. Specifically, we dispatch subproblems based on the object id. So each Check for a particular object will consistently land on the same peer, which will maximize cache efficiency for localized subproblems (e.g overlapping subproblems on the same object id through computed usersets).

2. When processing a DispatchCheck RPC, the `dispatchV1Service`, which implements the `DispatchService` service definition, we delegate to local check resolution via a `localDispatcher` which implements local graph evaluation. The delegate for the `localDispatcher` is a `cachedLocalDispatcher` which will serve a Check subproblem immediately from the cache if the subproblem has been computed recently. Since we've landed on a particular peer to resolve the local graph evaluation, then we can assume that the peer handling the local dispatch is the owner of the consistent hash key that the Check request hashes to using the consistent hashring.

3. When processing a DispatchCheck RPC, if the request is not cached already, then we redispatch the request. Dispatching a request incurrs a network hop to the peer which owns that object id, but this cost is no more expensive then doing an external cache lookup before doing further evaluation anyways. No network hops are necessary if the peer serving the DispatchCheck has the subproblem cached already (because it owns the cache key in the hashring).

## Getting Started

### Prerequisites
To run the sample code you need to spin up a Kubernetes cluster. We use Kubernetes headless services for service discovery between FGA servers. The easiest way to do this is with `minikube`.

```shell
minikube start
eval $(minikube docker-env) # exposes host docker images to the Kubernetes cluster in minikube
```

### Building the image
Build the image with `docker build -t dispatcher:0.0.1 .`

### Deploy the FGA services
```shell
kubectl apply -f k8s.yaml
```

Verify the pods are running (there should be 2 pods).
```shell
kubectl get pods

NAME                 READY   STATUS    RESTARTS   AGE
fga-b95d5b57-dh27z   1/1     Running   0          30s
fga-b95d5b57-rh9vs   1/1     Running   0          30s
```

### Port-forward to the FGAService server
```shell
kubectl port-forward svc/fga 50051:50051
```
> ℹ️ Kubernetes actually attaches/binds to a single pod when you port-forward. Any network traffic against port 50051 will be directed to a _singular_ pod.

### Execute an FGAService/Check RPC
```shell
grpcurl -v --plaintext -use-reflection -d '{"object_type": "document", "object_id": "1", "relation": "viewer"}' localhost:50051 fga.v1.FGAService/Check
```

### Tail the Logs
```shell
kubectl logs -f -l app.kubernetes.io/name=fga --prefix=true

1 [pod/fga-b95d5b57-dh27z/fga] Check has been called
2 [pod/fga-b95d5b57-dh27z/fga] cachedDispatcher.DispatchCheck (document:1#viewer)
3 [pod/fga-b95d5b57-dh27z/fga] cache miss 'document:1#viewer'
4 [pod/fga-b95d5b57-dh27z/fga] calling peer.DispatchClient.DispatchCheck 'document:1#viewer' - this goes over the network
5 [pod/fga-b95d5b57-dh27z/fga] DispatchCheck response 'true'
6 [pod/fga-b95d5b57-rh9vs/fga] (document:1#viewer) - dispatching document:1#editor
7 [pod/fga-b95d5b57-rh9vs/fga] cachedDispatcher.DispatchCheck (document:1#editor)
8 [pod/fga-b95d5b57-rh9vs/fga] serving 'document:1#editor' from cache
9 [pod/fga-b95d5b57-rh9vs/fga] 8.417µs
```
> ℹ️ You aren't guaranteed to get the logs for each pod in order relative to one another, but all logs for a specific pod (e.g. `pod/fga-b95d5b57-rh9vs/fga`) will be ordered. So if the logs are out of order relative to the output above, then that's not abnormal.

The log output above demonstrates how the dispatching is working. Here are some key take aways:

1. Pod `pod/fga-b95d5b57-dh27z/fga` receives the Check request (line 1). The first thing it does (line 2) is check if the subproblem is cached (if it is then we assume that pod owns the cache key). In this case, we see that there is a cache miss for `document:1#viewer` (line 3), so we proceed to dispatch the problem to the appropriate peer (line 4).

2. When pod `pod/fga-b95d5b57-dh27z/fga` dispatches the `document:1#viewer` problem to the peer dispatcher this invokes the DispatchCheck RPC on the DispatchServiceClient gRPC client (line 4) which chooses the peer to dispatch the RPC to based on a consistent hashring. We find the peer which owns the cache key for `document:1` by using Kubernetes headless service discovery and dispatch to that peer.

3. Pod `pod/fga-b95d5b57-rh9vs/fga` receives the peer's request to resolve `document:1#viewer` and immediately checks to see if it has that subproblem cached. In this case it doesn't have that subproblem cached, so it resolves using a local dispatcher and finds that `document:1#viewer` is rewritten to `document:1#editor` (line 6), and so the peer looks up in its cache if `document:1#editor` is already resolved (line 7). The peer serving the dispatched request finds `document:1#editor` in the cache and immediately returns the result (line 8), which avoids another dispatch over the network.

4. With the response from the dispached subproblem, `pod/fga-b95d5b57-dh27z/fga` is now able to respond to the initial client's Check request (line 5).