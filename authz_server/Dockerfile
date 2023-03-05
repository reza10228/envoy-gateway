FROM golang:1.13 as build

ENV GO111MODULE=on

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o grpc_server .
FROM gcr.io/distroless/base

COPY --from=build ["/build/grpc_server", "/build/.env", "/"]

EXPOSE 50051

ENTRYPOINT ["/grpc_server"]
