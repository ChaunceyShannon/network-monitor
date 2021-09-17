FROM golang:1.17.1-buster as builder

ARG BIN_NAME=network-monitor

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY *.go ./

RUN go mod tidy
RUN go build -o ${BIN_NAME}

# FROM ubuntu:20.04

# RUN mkdir -p /app;\
#   apt update; \
#   apt install -y ca-certificates

FROM gcr.io/distroless/static:nonroot

ARG BIN_NAME=network-monitor

WORKDIR /app

COPY --from=builder /app/${BIN_NAME} /bin/${BIN_NAME}

CMD ["/bin/${BIN_NAME}", "-c", "/app/${BIN_NAME}.ini"]
