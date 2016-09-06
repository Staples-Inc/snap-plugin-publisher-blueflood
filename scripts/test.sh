TEST_DIRS="github.com/Staples-Inc/snap-plugin-publisher-blueflood/blueflood"
set -e
go test $TEST_DIRS -v -covermode=count
