cat > /etc/systemd/system/baxx-watcher.service <<EOL
[Unit]
Description=baxx watcher
After=postgres.service
Requires=docker.service postgres.service

[Service]
ExecStart=/baxx/scripts/prod/baxx/start-watcher.sh
Restart=always
RestartSec=60s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL


cat > /etc/systemd/system/baxx-status.service <<EOL
[Unit]
Description=baxx status
After=postgres.service
Requires=docker.service postgres.service

[Service]
ExecStart=/baxx/scripts/prod/baxx/start-status.sh
Restart=always
RestartSec=60s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable baxx-watcher
systemctl enable baxx-status

