FROM ubuntu
ENV GOPATH=/baxx/
RUN apt-get update && apt-get -y upgrade && apt-get install -y mdadm ca-certificates
ADD bin /baxx
ADD t /baxx/src/github.com/jackdoe/baxx/help/t

