#!/usr/bin/env python

import grpc
import activate_pb2_grpc as pb2_grpc
import activate_pb2 as pb2
from google.protobuf.struct_pb2 import Struct
import sys


class OptimizelyServiceClient(object):
    """
    Client for gRPC interface for optimizely-service
    """

    def __init__(self, host, port):
        self.host = host
        self.port = port

        # instantiate a channel
        self.channel = grpc.insecure_channel(
            f"{self.host}:{self.port}"
        )

        # bind the client and the server
        self.stub = pb2_grpc.ActivateStub(self.channel)

    
    def activate(self, experiment_key, user_id, attributes):
        attrs = Struct()
        for k, v in attributes.items():
            attrs.update({k: v})
        user = pb2.User(id=user_id, attributes=attrs)
        activate_request = pb2.ActivateRequest(experiment_key=experiment_key, user=user)
        return self.stub.Activate(activate_request).variation
        
    
if __name__ == "__main__":
    _, host, port = sys.argv
    client = OptimizelyServiceClient(host, port)
    attrs = {
        "customer_uuid": "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
        "country": "US",
        "public_id": "jhds"
    }
    experiment_key = "us-widget-bff"
    user_id = "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74"
    variation = client.activate(experiment_key, user_id, attrs)
    print(f"Variation: {variation}")


# experimentKey := "us-widget-bff"
# 	m := map[string]interface{}{
# 		"customer_uuid": "b5aedcf2-c5d8-4bd1-a4df-4d76702cea74",
# 		"country":       "US",
# 		"platform":      "mobile",
# 		"public_id":     "jhds",
# 	}