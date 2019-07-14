#!/bin/sh

baxx_put() {
    if [ $# -lt 2 ]; then
        echo "usage: $0 file dest [force]"
    else
        file=$1
        dest=$2
        force=${3:-noforce}
        sha=$(shasum -a 256 $file | cut -f 1 -d ' ')
        (curl -s https://baxx.dev/sync/sha256/$BAXX_TOKEN/$sha -f >/dev/null 2>&1 \
             && [[ "$force" != "force" ]] \
             && echo SKIP $file .. already baxxed, use \"$0 $1 $2 force\" to force) || \
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

baxx_delete() {
    if [ $# -ne 1 ]; then
        echo "usage: $0 file"
    else
        file=$1
        curl -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$file
    fi
}

baxx_rmdir() {
    if [ $# -ne 1 ]; then
        echo "usage: $0 path"
    else
        file=$1
        curl -d '{"force":true}' -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$file
    fi
}

baxx_rmrf() {
    if [ $# -ne 1 ]; then
        echo "usage: $0 path"
    else
        file=$1
        curl -d '{"force":true,"recursive":true}' -X DELETE https://baxx.dev/io/$BAXX_TOKEN/$file
    fi
}

baxx_ls() {
    curl https://baxx.dev/ls/$BAXX_TOKEN/$*
}

baxx_sync() {
    if [ $# -ne 1 ]; then
        echo "usage: $0 path"
    else
        find $1 -type f \
            | xargs -P4 -I '{}' \
                    shasum -a 256 {} \
            | curl -s --data-binary @- https://baxx.dev/sync/sha256/$BAXX_TOKEN \
            | awk '{ print $2 }' \
            | xargs -P4 -I '{}' \
                    curl -s -T {} https://baxx.dev/io/$BAXX_TOKEN/backup/{}
    fi
}


baxx_share() {
    if [ $# -lt 1 ]; then
        echo "usage: $0 file"
    else
        fn=$1
        ct=$(file --mime-type $fn | cut -f 2 -d ' ')
        curl -H "Content-Type: $ct" -s \
             -T $file https://baxx.dev/io/$BAXX_TOKEN/public/$fn > /dev/null && \
            curl -s -XPOST https://baxx.dev/share/$BAXX_TOKEN/public/$fn | \
                grep link | cut -f 4 -d '"'
    fi
}

baxx_paste() {
    fn=paste.$(date +%s)
    (curl -H "Content-Type: text/plain; charset=utf-8" -XPOST -s --data-binary @- \
          https://baxx.dev/io/$BAXX_TOKEN/public/$fn > /dev/null && \
         curl -s -XPOST https://baxx.dev/share/$BAXX_TOKEN/public/$fn | grep link \
             | cut -f 4 -d '"')
}
