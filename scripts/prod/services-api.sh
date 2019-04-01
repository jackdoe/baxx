cat > /etc/systemd/system/baxx-api.service <<EOL
[Unit]
Description=baxx api
After=judoc.service postgres.service scylla.service
Requires=docker.service judoc.service postgres.service scylla.service

[Service]
ExecStart=/baxx/scripts/prod/baxx/start-api.sh
Restart=always
RestartSec=10s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL

cat > /etc/systemd/system/baxx-email-send.service <<EOL
[Unit]
Description=baxx email send
After=postgres.service
Requires=docker.service postgres.service

[Service]
ExecStart=/baxx/scripts/prod/baxx/start-email-send.sh
Restart=always
RestartSec=60s
Type=simple

[Install]
WantedBy=multi-user.target
EOL


cat > /etc/systemd/system/baxx-notification-rules.service <<EOL
[Unit]
Description=baxx notification rules
After=postgres.service
Requires=docker.service postgres.service

[Service]
ExecStart=/baxx/scripts/prod/baxx/start-notification-rules.sh
Restart=always
RestartSec=60s
Type=simple

[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable baxx-api
systemctl enable baxx-email-send
if [ $(hostname) = "bb.baxx.dev" ]; then
    systemctl enable baxx-notification-rules
fi
