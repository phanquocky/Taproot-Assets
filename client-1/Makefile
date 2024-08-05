create_wallet:
	export WPORT=$(wport)
	./create_wallet.sh $(username) $(wport)


create_addr:
	export WPORT=$(wport)
	ADDRESS=$$(btcctl --simnet --notls -C "" --wallet --rpcserver 127.0.0.1:$(wport) --rpcuser=admin --rpcpass=admin123 getnewaddress); \
	echo "ADDRESS: $$ADDRESS"; \
	echo "\nWPORT=$(wport)\n" >> .env; \
	echo "WADDR=$$ADDRESS" >> .env; \

build:
	go build -o client_side main.go
