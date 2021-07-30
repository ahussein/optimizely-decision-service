FROM golang:1.13.0-alpine AS builder

WORKDIR /app

COPY . .

ARG GITHUB_TOKEN

ENV CGO_ENABLED 0
ENV GOOS linux

RUN apk add make git

RUN make all

# --- END OF RAML BUILDER

# APP IMAGE
FROM ubuntu:bionic

COPY --from=builder /app/build/optimizely-decision-service-api /optimizely-decision-service-api
EXPOSE 80
CMD ["optimizely-decision-service-api"]