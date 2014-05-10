FROM ubuntu:precise
RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -qy --fix-missing build-essential curl git mercurial
RUN curl -s https://go.googlecode.com/files/go1.2.1.src.tar.gz | tar -v -C /usr/local -xz
RUN cd /usr/local/go/src && ./make.bash --no-clean 2>&1
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /opt/go
ADD . /opt/go/src/github.com/secondbit/gifs/api
RUN cd /opt/go/src/github.com/secondbit/gifs/api/gifsd && go get . && go build .
EXPOSE 8080
ENTRYPOINT ["/opt/go/src/github.com/secondbit/gifs/api/gifsd/gifsd", "-etcd-address=http://172.17.42.1:4001"]
