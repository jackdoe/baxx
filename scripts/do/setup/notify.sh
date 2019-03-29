apt update
apt upgrade -y
apt install -y postgresql-client-10
snap install go --classic

adduser baxx --home /home/baxx --disabled-password --gecos ""
passwd -l baxx

go get github.com/jackdoe/baxx/...

cat > /etc/systemd/system/send_email_queue.service <<EOF
[Unit]
Requires=
After=

[Service]
Type=simple
User=baxx
ExecStart=/home/baxx/send_email_queue -debug
Restart=on-failure
EnvironmentFile=/home/baxx/config
[Install]
WantedBy=multi-user.target
EOF

cat > /home/baxx/config <<EOF
BAXX_POSTGRES='host=localhost port=5432 user=baxx dbname=baxx password=baxx'
BAXX_SENDGRID_KEY='XXX'
BAXX_SLACK_PANIC='XXX'
EOF

chown root /home/baxx/config
chmod 600 /home/baxx/config
systemctl daemon-reload
systemctl enable send_email_queue
service send_email_queue start

# crontab 1 14 * * * root source /home/baxx/config && /home/baxx/notification_run -debug
