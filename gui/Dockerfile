FROM golang:alpine

RUN apk update && \
    apk add bash openssh git && \
    deluser $(getent passwd 33 | cut -d: -f1) && \
    delgroup $(getent group 33 | cut -d: -f1) 2>/dev/null || true && \
    cp -a /etc/ssh /etc/ssh.cache && \
    rm -rf /var/cache/apk/*

ADD docker/sshd_config /etc/ssh/sshd_config
ADD src /src
RUN go get -v github.com/jackdoe/baxx/client github.com/jackdoe/baxx/common github.com/jackdoe/baxx/help github.com/marcusolsson/tui-go
RUN mkdir /gui && cd /src && go build -o /gui/baxx.gui

# dont docker push
RUN ssh-keygen -A

EXPOSE 22

COPY docker/entry.sh /entry.sh

ENTRYPOINT ["/entry.sh"]

CMD ["/usr/sbin/sshd", "-D", "-f", "/etc/ssh/sshd_config"]