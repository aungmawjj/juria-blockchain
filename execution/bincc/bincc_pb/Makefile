
all: bincc.pb.go

bincc.pb.go: bincc.proto
	protoc \
	--go_out=./ \
	--go_opt=Mexecution/bincc/bincc_pb/bincc.proto=github.com/aungmawjj/juria-blockchain/execution/bincc/bincc_pb \
	bincc.proto

clean:
	rm -f *.pb.go
