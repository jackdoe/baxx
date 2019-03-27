todo_set() {
    curl -f -s --data-binary @- https://baxx.dev/io/$BAXX_TOKEN/todo.txt > /dev/null
}

todo_get() {
    curl https://baxx.dev/io/$BAXX_TOKEN/todo.txt -f -s || (echo -n | todo_set > /dev/null)
}

todo_add() {
    { todo_get && echo [ $(date) ] $* } | todo_set && todo_get
}

todo_done() {
    (
        ITEM=$*
        todo_get | \
            while read -r line
            do
                case "$line" in
                    *$ITEM) echo $line [ $(date) ] ;;
                    *) echo $line ;;
                esac
            done
    ) | todo_set && todo_get
}

todo_delete_matching() {
    (
        ITEM=$*
        todo_get | \
            while read -r line
            do
                case "$line" in
                    *$ITEM*) ;;
                    *) echo $line ;;
                esac
            done
    ) | todo_set && todo_get
}