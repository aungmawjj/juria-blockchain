
all: core.pb.go

core.pb.go: core.proto
	protoc \
	--proto_path=../..:. \
	--go_out=./ \
	--go_opt=Mcore/core_pb/core.proto=github.com/aungmawjj/juria-blockchain/core/core_pb \
	core.proto

clean:
	rm -f *.pb.go
