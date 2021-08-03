# optimizely-decision-service
A proxy microservice for the Optimizely Golang SDK


protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--go-grpc_opt=require_unimplemented_servers=false \
internal/activate/activate.proto
	    
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