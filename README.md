# optimizely-decision-service
A proxy microservice for the Optimizely Golang SDK


protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--go-grpc_opt=require_unimplemented_servers=false \
internal/activate/activate.proto
	    
## Kubectl port forward command
kubectl port-forward svc/optimizely-decision-service-grpc-service 8080:80

## Run the go client
go run cmd/client/main.go