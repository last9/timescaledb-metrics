#!/bin/sh
set -e

RUN=$1
FLAGS=$2

set +ex

check_errcheck() {
  errcheck ${FLAGS} ./...
}

check_fmt() {
  gofmt -l .
}

check_imports() {
  goimports -l .
}

check_lint() {
  golint -set_exit_status ./...
}

check_sec() {
  which gosec || go install -v github.com/securego/gosec/v2/cmd/gosec
  gosec -out result.txt ${FLAGS} ./...
}

check_shadow() {
  which shadow || go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow
  go vet -vettool=`which shadow` ${FLAGS} ./...
}

check_staticcheck() {
  which staticcheck || go install -v honnef.co/go/tools/cmd/staticcheck
  staticcheck ${FLAGS} ./...
}

check_vet() {
  go vet ${FLAGS} ./...
}

case ${RUN} in
	"errcheck" )
		check_errcheck
		;;
	"fmt" )
		check_fmt
		;;
	"imports" )
		check_imports
		;;
	"lint" )
		check_lint
		;;
	"sec" )
		check_sec
		;;
	"shadow" )
		check_shadow
		;;
	"staticcheck" )
		check_staticcheck
		;;
	"vet" )
		check_vet
		;;
	* )
    check_imports
    check_fmt
    check_lint
    check_sec
    check_shadow
    check_errcheck
    check_staticcheck
    check_vet
esac
