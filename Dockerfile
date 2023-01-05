FROM golang:1.19-alpine as builder
WORKDIR /go/src/github.com/dnachev/wg-nc
RUN apk add make
ADD . ./
RUN make build-alpine

FROM scratch
COPY --from=builder /go/src/github.com/dnachev/wg-nc/bin/wg-nc /wg-nc

ENTRYPOINT ["./wg-nc"]
