#!/bin/bash

source /root/.bashrc

curl https://baxx.dev/io/$BAXX_TOKEN/letsencrypt.tar | \
    encrypt -k /root/.pw -d | \
    tar -xf - -C /

echo ok | curl -s -d@- "https://baxx.dev/io/$BAXX_TOKEN/letsencrypt-downloaded-bb.monitor?age=3600"