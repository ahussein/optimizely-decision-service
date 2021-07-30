FROM golang:1.13.0-alpine AS builder

WORKDIR /app

COPY . .

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add make git

RUN make build-grpc

RUN ls /app/build

# APP IMAGE
FROM ubuntu:bionic

RUN apt-get update -y && apt-get install -y curl 
WORKDIR /app
COPY --from=builder /app/build/optimzely-decision-service-grpc /app/optimzely-decision-service-grpc
EXPOSE 50051
CMD ["/app/optimzely-decision-service-grpc"]