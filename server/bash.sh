#!/bin/zsh

#ip_address=$(ifconfig | awk '/inet /{print $2}' | grep -oE "192\.168\.[0-9]{1,3}\.[0-9]{1,3}")
ip_address=$(ifconfig | awk '/inet /{print $2}' | grep -oE "172\.20\.[0-9]{1,3}\.[0-9]{1,3}")
echo $ip_address

echo "Creating btcd"
~/go/bin/btcd --simnet --rpcuser=admin --rpcpass=admin123 &
btcdid=$(pgrep -o btcd)

sleep 2

BTCPATH=~/Library/Application\ Support

rm -r "$BTCPATH/Btcwallet/simnet"

echo "Creating btcwallet"
~/go/bin/btcwallet --simnet --create

~/go/bin/btcwallet --simnet --username=admin --password=admin123 &
btcwalletid=$(pgrep -o btcwallet)

sleep 2

miningaddr=`~/go/bin/btcctl --simnet --wallet --rpcuser=admin --rpcpass=admin123 getnewaddress`

echo "Mining address: "
echo $miningaddr

echo "Kill old btcd btcwallet: "
sudo kill $btcwalletid
sudo kill $btcdid

echo "remove rpc key"
rm "$BTCPATH/Btcd/rpc.cert" "$BTCPATH/Btcd/rpc.key"

sleep 2

~/go/bin/btcd --simnet --rpcuser=admin --rpcpass=admin123 --txindex --rpclisten="$ip_address:18556" --miningaddr=$miningaddr &

sleep 2

rm "$BTCPATH/Btcwallet/rpc.cert" "$BTCPATH/Btcwallet/rpc.key" "$BTCPATH/Btcwallet/btcd.cert"

cp "$BTCPATH/Btcd/rpc.cert" "$BTCPATH/Btcwallet/btcd.cert"

echo "Running btcwallet"
~/go/bin/btcwallet --simnet --username=admin --password=admin123 --rpcconnect="$ip_address:18556" &

sleep 2

echo "generate token by btcctl"
~/go/bin/btcctl --simnet --wallet --rpcuser=admin --rpcpass=admin123 generate 100

echo $ip_address
echo $miningaddr