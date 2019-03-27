_todo_set() {
    curl -f -s --data-binary @- https://baxx.dev/io/$BAXX_TOKEN/todo.txt > /dev/null
}

todo_list() {
    curl https://baxx.dev/io/$BAXX_TOKEN/todo.txt -f -s || (echo -n | _todo_set > /dev/null)
}

todo_add() {
    if [ $# -eq 0 ]; then
        echo "usage: $0 item"
    else
        { todo_list && echo [ $(date) ] $* } | _todo_set && todo_list
    fi
}

todo_done() {
    if [ $# -eq 0 ]; then
        echo "usage: $0 item"
    else
        (
            ITEM=$*
            todo_list | \
                while read -r line
                do
                    case "$line" in
                        *$ITEM) echo $line [ $(date) ] ;;
                        *) echo $line ;;
                    esac
                done
        ) | _todo_set && todo_list
    fi
}

todo_delete_matching() {
    if [ $# -eq 0 ]; then
        echo "usage: $0 item"
    else
        (
            ITEM=$*
            todo_list | \
                while read -r line
                do
                    case "$line" in
                        *$ITEM*) ;;
                        *) echo $line ;;
                    esac
                done
        ) | _todo_set && todo_list
    fi
}
