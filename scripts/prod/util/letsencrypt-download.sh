#!/bin/sh

curl https://baxx.dev/io/$BAXX_TOKEN/letsencrypt.tar | \
    encrypt -k /root/.pw -d | \
    tar -xf - -C /

