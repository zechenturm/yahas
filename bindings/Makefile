GOPLUG = go build -buildmode=plugin

SRCS := $(wildcard *.go)
PLUGNAMES := $(SRCS:%.go=%)
SOS  := $(PLUGNAMES:%=%.so)

%.so: %.go
	$(GOPLUG) $^

all: $(SOS)

clean:
	rm $(SOS)

.PHONY: all