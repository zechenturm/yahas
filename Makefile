
NMAKE := make -j$(shell nproc)

SRCS := $(wildcard *.go)


.PHONY: all
.PHONY: bindings
.PHONY: bindings-clean
.PHONY: plugins
.PHONY: plugins-clean
.PHONY: run
.PHONY: yahas
.PHONY: clean

all: bindings plugins yahas

bindings:
	$(NMAKE) -C bindings

bindings-clean:
	make -C bindings clean

plugins:
	$(NMAKE) -C plugins

plugins-clean:
	make -C plugins clean

yahas: $(SRCS) plugins bindings
	go build -o yahas

run: yahas
	./yahas

clean: bindings-clean plugins-clean
	rm yahas