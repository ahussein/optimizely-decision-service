# optimizely-decision-service
`optimizely-decision-service` expose gRPC services to provide all the features in the Optimizely SDK to consumers. It wraps the Golang SDK.
## generate go client
```
brew install protobuf

go get -u github.com/golang/protobuf/protoc-gen-go
```
go to `<PROJECT_DIR>/cmd/grpc` and then execute

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/experiment.proto
```

## generate python client
### install python tools
```
pip3 install grpcio grpcio-tools
```
### generate client code
```
python3 -m grpc_tools.protoc --proto_path=. ./internal/activate/activate.proto --python_out=cmd/client/python/ --grpc_python_out=cmd/client/python
```
	    
## Kubectl port forward command
kns default
kubectl port-forward svc/optimizely-decision-service-grpc-service 8080:80

## Run the go client
go run cmd/client/main.go

## Run the python client
python3 cmd/client/python/client.py localhost 8080

## call optimizely service

```
curl -i -X "POST" "https://optimizely-service.staging-k8s.hellofresh.io/activate" \
     -H 'Content-Type: application/json; charset=utf-8' \
     -d $'{
  "experiment_key": "us-widget-bff",
  "user_id": "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
  "attributes": {
    "customer_uuid": "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
    "country": "US",
    "public_id": 101
  }
}'
```