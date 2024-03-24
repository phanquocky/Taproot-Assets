#!/bin/zsh

~/go/bin/btcd --simnet --rpcuser=admin --rpcpass=admin123 &
btcdid=$(pgrep -o btcd)

sleep 2

BTCPATH=~/Library/Application\ Support

rm -r "$BTCPATH/Btcwallet/simnet"

~/go/bin/btcwallet --simnet --create

~/go/bin/btcwallet --simnet --username=admin --password=admin123 &

sleep 2

miningaddr=`~/go/bin/btcctl --simnet --wallet --rpcuser=admin --rpcpass=admin123 getnewaddress`

echo "Mining address: "
echo $miningaddr

sudo kill -9 $btcdid

~/go/bin/btcd --simnet --rpcuser=admin --rpcpass=admin123 --txindex --miningaddr=$miningaddr &

sleep 2
echo $miningaddr