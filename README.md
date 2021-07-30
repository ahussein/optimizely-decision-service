# optimizely-decision-service
A proxy microservice for the Optimizely Golang SDK


protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
--go-grpc_opt=require_unimplemented_servers=false \
internal/activate/activate.proto
	    
