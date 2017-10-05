FROM golang:1.9.0
WORKDIR /go/src/github.com/IBM/ubiquity-docker-plugin/
COPY . .
RUN go get -v github.com/Masterminds/glide \
&& glide up \
&& CGO_ENABLED=1 GOOS=linux go build -tags netgo -v -a --ldflags '-w -linkmode external -extldflags "-static"' -installsuffix cgo -o ubiquity-docker-plugin main.go


FROM alpine:latest
RUN apk --update add ca-certificates multipath-tools nfs-utils open-iscsi openssh sg3_utils \
&& mkdir -p /ubiquity /run/docker/plugins
WORKDIR /root
COPY --from=0 /go/src/github.com/IBM/ubiquity-docker-plugin/ubiquity-docker-plugin .
CMD ["/root/ubiquity-docker-plugin"]
