
.PHONY: all
all:
	mkdir -p ./bindings
	rm -rf bindings
	mkdir -p ./bindings
	mkdir -p solc-build
	rm -rf ./solc-build
	mkdir -p solc-build
	solc8 --abi contract/P2E.sol -o solc-build
	solc8 --bin contract/P2E.sol -o solc-build --optimize
	abigen --bin=solc-build/P2E.bin --abi=solc-build/P2E.abi --pkg=bindings --out=bindings/P2E.go --type=P2EContract
	abigen --abi=solc-build/IBEP20.abi --pkg=bindings --out=bindings/IBEP20.go --type=IBEP20
	cp ./solc-build/P2E.abi ./fe/abi.json
	go get
	#go build -gcflags="all=-N -l" server.go # uncomment for debugger support
	go build server.go