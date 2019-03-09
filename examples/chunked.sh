#!/bin/bash

baxx_chunk_put() {
    PID=$$
    if [ $# -ne 2 ]; then
        echo "usage: $0 file dest"
    else
        file=$1
        dest=$2
        split -n 5 $file baxx.$PID.
        for ext in aa ab ac ad ae; do
            part=baxx.$PID.$ext
            sha=$(shasum -a 256 $part | cut -f 1 -d ' ')
            (curl -s https://baxx.dev/sha256/$BAXX_TOKEN/$sha -f > /dev/null 2>&1  && echo -n "skipping $file $part") || \
                (echo -n "$file $part .." ; curl -T $part https://baxx.dev/io/$BAXX_TOKEN/chunked/$dest.$ext && rm $part && echo -n .. [ done ])
            echo
        done 
    fi
}

baxx_chunk_get() {
    if [ $# -ne 2 ]; then
        echo "usage: $0 file dest"
    else
        file=$1
        dest=$2
        :>$dest
        for ext in aa ab ac ad ae; do
            echo -n $file.$ext ..
            curl -s https://baxx.dev/io/$BAXX_TOKEN/chunked/$file.$ext >> $dest
            echo -n .. [ done ]
            echo
        done 
    fi
}

