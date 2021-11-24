FROM golang:alpine

# Dependencies
RUN apk add --update make bash libtool git python3 autoconf automake g++ boost-dev busybox-extras curl

# Add code to gopath and build
RUN mkdir -p src/github.com/algorand/indexer
WORKDIR src/github.com/algorand/indexer
COPY . .
RUN make

# Launch indexer with a script
COPY run.sh /tmp/run.sh
CMD ["/tmp/run.sh"]