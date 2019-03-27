# Example of having a todo list file stored in baxx.dev
#
#   % todo_add buy milk
#   [ Wed 27 Mar 21:10:50 CET 2019 ] buy milk
#   % todo_add fix milk
#   [ Wed 27 Mar 21:10:50 CET 2019 ] buy milk
#   [ Wed 27 Mar 21:10:57 CET 2019 ] fix milk
#   % todo_done fix milk
#   [ Wed 27 Mar 21:10:50 CET 2019 ] buy milk
#   [ Wed 27 Mar 21:10:57 CET 2019 ] fix milk [ Wed 27 Mar 21:11:04 CET 2019 ]
#   % todo_delete_matching fix milk
#   [ Wed 27 Mar 21:10:50 CET 2019 ] buy milk
#   %

TODOFILE=todo.txt

_todo_set() {
    curl -f -s --data-binary @- https://baxx.dev/io/$BAXX_TOKEN/$TODOFILE > /dev/null
}

todo_list() {
    curl https://baxx.dev/io/$BAXX_TOKEN/$TODOFILE -f -s || (echo -n | _todo_set > /dev/null)
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
