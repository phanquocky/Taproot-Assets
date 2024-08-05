build:
	export WPORT=$(wport)
	./create_wallet.sh $(username) $(wport)
	go build -o $(username) main.go

create_addr:
	export WPORT=$(wport)
	@ADDRESS=$$(btcctl --simnet --notls -C "" --wallet --rpcserver 127.0.0.1:$(wport) --rpcuser=admin --rpcpass=admin123 getnewaddress)
	echo "$$ADDRESS"
