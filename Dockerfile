FROM golang:1.18.7 as builder
WORKDIR /go/src/github.com/dddpaul/gonc
ADD . ./
RUN make build-alpine

FROM alpine:latest
WORKDIR /app
COPY --from=builder /go/src/github.com/dddpaul/gonc/bin/gonc .

ENTRYPOINT ["./gonc"]
