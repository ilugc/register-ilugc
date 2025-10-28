all: register

register:
	(cd src/cmd/; CGO_ENABLED=0 go build -ldflags '-s -w' -o register)

clean:
	-rm src/cmd/register

.PHONY: clean
