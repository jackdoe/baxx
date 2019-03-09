#!/bin/sh

baxx_put() {
    if [ $# -ne 2 ]; then
        echo "usage: $0 file dest"
    else
        file=$1
        dest=$2
        curl -T $file https://baxx.dev/io/$BAXX_TOKEN/$dest
    fi
}

baxx_get() {
    if [ $# -ne 2 ]; then
        echo "usage: $0 file dest"
    else
        file=$1
        dest=$2
        curl -s https://baxx.dev/io/$BAXX_TOKEN/$file > $dest 
    fi
}

baxx_ls() {
    curl https://baxx.dev/ls/$BAXX_TOKEN/
}
