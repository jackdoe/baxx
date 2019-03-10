#!/bin/sh

baxx_put() {
    if [ $# -lt 2 ]; then
        echo "usage: $0 file dest [force]"
    else
        file=$1
        dest=$2
        force=${3:-noforce}
        sha=$(shasum -a 256 $file | cut -f 1 -d ' ')
        (curl -s https://baxx.dev/sha256/$BAXX_TOKEN/$sha -f >/dev/null 2>&1 \
             && [[ "$force" != "force" ]] \
             && echo SKIP $file .. already baxxed use \"$0 $1 $2 force\" to force) || \
            curl -T $file https://baxx.dev/io/$BAXX_TOKEN/$dest
    fi
}

baxx_get() {
    if [ $# -ne 2 ]; then
        echo "usage: $0 file dest"
    else
        file=$1
        dest=$2
        curl https://baxx.dev/io/$BAXX_TOKEN/$file > $dest
    fi
}


baxx_ls() {
    curl https://baxx.dev/ls/$BAXX_TOKEN/$*
}

