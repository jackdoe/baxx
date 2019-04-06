#!/bin/bash
source /root/.bashrc

tar -cvf - /etc/letsencrypt | \
    encrypt -k /root/.pw | \
    curl --data-binary @- "https://baxx.dev/io/$BAXX_TOKEN/letsencrypt.tar?age=3600&delta=80"
