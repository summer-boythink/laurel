BLDDIR = build
APPS = pages
DB_FILE = run_test.db
CMD_DIR = cmd

all: build

build: clean fmt
	mkdir -p $(BLDDIR)
	go build -o $(BLDDIR)/$(APPS) ./$(CMD_DIR)/$(APPS)

run: build
	./$(BLDDIR)/$(APPS) $(DB_FILE)

clean:
	rm -fr $(BLDDIR)

fmt:
	go mod tidy
	gofmt -w .

test:
	go test -C ./$(CMD_DIR)/$(APPS) -v

.PHONY: all clean fmt test
