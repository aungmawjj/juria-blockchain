PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

%.pb.go: %.proto
		protoc --proto_path=$(PWD)/..:. --go_out=$(PWD) $<

clean:
		rm -f *.pb.go
		rm -f *.go
