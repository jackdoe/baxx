if [ $(hostname) = "bb.baxx.dev" ]; then
    ################################
    # MASTERMASTERMASTERMASTERMAST #
    ################################

    cat > /etc/systemd/system/postgres.service <<EOL
[Unit]
Description=Postgres
After=docker.service
Requires=docker.service

[Service]
ExecStart=/baxx/scripts/prod/postgres/master/start.sh
Restart=always
RestartSec=10s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL
else
    ################################
    # REPLICAREPLICAREPLICAREPLICA #
    ################################
    cat > /etc/systemd/system/postgres.service <<EOL
[Unit]
Description=Postgres
After=docker.service
Requires=docker.service

[Service]
ExecStart=/baxx/scripts/prod/postgres/slave/start.sh
Restart=always
RestartSec=10s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL

fi


cat > /etc/systemd/system/scylla.service <<EOL
[Unit]
Description=Scylla
After=docker.service
Requires=docker.service

[Service]
ExecStart=/baxx/scripts/prod/scylla/start.sh
Restart=always
RestartSec=10s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL

cat > /etc/systemd/system/judoc.service <<EOL
[Unit]
Description=Judoc

After=docker.service scylla.service
Requires=docker.service scylla.service

[Service]
ExecStart=/baxx/scripts/prod/judoc/start.sh
Restart=always
RestartSec=10s
Type=simple
 
[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable postgres
systemctl enable scylla
systemctl enable judoc






