
default:
	$(MAKE) deps
	$(MAKE) all
deps:
	bash -c "glide install"
test:
	bash -c "./scripts/test.sh $(TEST)"
all:
	bash -c "./scripts/build.sh $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))"