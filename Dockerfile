FROM golang:alpine AS builder
LABEL maintainer="Mikołaj Gąsior"

RUN apk add --update git bash openssh make gcc musl-dev

WORKDIR /go/src/mikolajgasior/octo-linter
COPY . .
RUN cd cmd/octo-linter && go generate && go build -o octo-linter

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /bin
COPY --from=builder /go/src/mikolajgasior/octo-linter/cmd/octo-linter/octo-linter octo-linter
RUN chmod +x /bin/octo-linter
RUN /bin/octo-linter
ENTRYPOINT ["/bin/octo-linter"]
