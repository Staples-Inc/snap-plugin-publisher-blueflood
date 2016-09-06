TEST_DIRS="github.com/intelsdi-x/snap-plugin-publisher-blueflood/blueflood"
set -e
go test $TEST_DIRS -v -covermode=count