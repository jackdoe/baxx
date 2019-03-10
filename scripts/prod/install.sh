
apt update
apt upgrade -y
apt install -y postgresql-client-10
snap install go --classic

adduser baxx --home /home/baxx --disabled-password --gecos ""
passwd -l baxx

go get github.com/jackdoe/baxx/...
cd go/src/github.com/jackdoe/baxx/api && go build -o /home/baxx/baxx.api

## FIXME(jackdoe) MOVE TO DOCKER

cat > /etc/systemd/system/baxx.service <<EOF
[Unit]
Requires=
After=

[Service]
Type=simple
User=baxx
ExecStart=/home/baxx/baxx.api -debug
Restart=on-failure
EnvironmentFile=/home/baxx/config
[Install]
WantedBy=multi-user.target
EOF

cat > /home/baxx/config <<EOF
BAXX_POSTGRES='host=localhost port=5432 user=baxx dbname=baxx password=baxx'
BAXX_S3_SECRET='XXX'
BAXX_S3_ACCESS_KEY='XXX'
BAXX_S3_TOKEN='XXX'
BAXX_S3_BUCKET='baxx'
BAXX_S3_ENDPOINT='ams3.digitaloceanspaces.com'
BAXX_S3_REGION='ams3'
BAXX_SENDGRID_KEY='XXX'
EOF

chown root /home/baxx/config
chmod 600 /home/baxx/config
systemctl daemon-reload
systemctl enable baxx
service baxx start
