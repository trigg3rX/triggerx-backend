FROM golang:1.21 as build

RUN apt update && \
    apt install git -y && \
    git clone https://github.com/samlaf/go-ethereum.git && \
    cd go-ethereum/cmd/abigen && \
    git checkout issue-28278/abigen-interfaces && \
    go build .

FROM debian:latest
COPY --from=build /go/go-ethereum/cmd/abigen/abigen /usr/local/bin/abigen
ENTRYPOINT [ "abigen"]
