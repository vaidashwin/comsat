# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang
ADD . /go/src/github.com/vaidashwin/comsat

# Vars
ENV discord_token ""

# Install stuff
WORKDIR /go/src/github.com/vaidashwin/comsat
RUN go get github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go install github.com/vaidashwin/comsat

# Pew pew gogogo
ENTRYPOINT /go/bin/comsat -token "$discord_token" -config /etc/comsat.json
EXPOSE 7547