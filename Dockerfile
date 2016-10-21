# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:onbuild

# Copy the local package files to the container's workspace.
#ADD . /go/src/github.com/traderboy/gorest

# Build the gorest command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
#CMD cd /go/src/github.com/traderboy/gorest; go get github.com/gin-gonic/gin && go build 
#CMD cd /go/src/github.com/traderboy/gorest; go get github.com/mattn/go-sqlite3 && go build


#RUN go install github.com/traderboy/gorest


# Run the outyet command by default when the container starts.
#ENTRYPOINT /go/bin/gorest

# Document that the service listens on port 8080.
#EXPOSE 8080