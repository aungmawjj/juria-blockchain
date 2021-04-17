
all: p2p.pb.go

p2p.pb.go: 
	protoc \
	--proto_path=../..:. \
	--go_out=./ \
	--go_opt=Mp2p/p2p_pb/p2p.proto=github.com/aungmawjj/juria-blockchain/p2p/p2p_pb \
	--go_opt=Mcore/core_pb/core.proto=github.com/aungmawjj/juria-blockchain/core/core_pb \
	p2p.proto

clean:
	rm -f *.pb.go
