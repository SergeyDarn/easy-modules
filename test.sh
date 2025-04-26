gotest() {
    go test $* | sed ''/PASS/s//$(printf "\033[\e[1;32mPASS\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[\e[1;31mFAIL\033[0m")/'' | sed ''/FAIL/s//$(printf "\033[\e[1;31mFAIL\033[0m")/'' | GREP_COLOR="01;31" egrep --color=always '\s*[a-zA-Z0-9\-_.]+[:][0-9]+[:].*|^'
}

ENV_ROOT=".." gotest -v ./...
