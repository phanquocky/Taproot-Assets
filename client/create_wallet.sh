#!/usr/bin/expect -f

# Get the filename from the command line arguments
set filename [lindex $argv 0]

set wport [lindex $argv 1]

set password "admin"

set wallet_db "wallet/$filename/simnet/wallet.db"

# Check if wallet database file exists and skip creation if it does
if { [file exists $wallet_db] } {
    puts "Wallet database file $wallet_db already exists. Skipping wallet creation."
    spawn ~/go/bin/btcwallet --simnet --noclienttls --noservertls -A "wallet/$filename" --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123 --rpcconnect 127.0.0.1:8000 --rpclisten 127.0.0.1:$wport
    interact
}

spawn ~/go/bin/btcwallet --create --simnet --noclienttls --noservertls -A "wallet/$filename" --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123 --rpcconnect 127.0.0.1:8000 --rpclisten 127.0.0.1:$wport

expect "Enter the private passphrase for your new wallet:" 
send "$password\r"

expect "Confirm passphrase:"
send "$password\r"

send "no\r"  ;# Send "no" followed by Enter to opt out of additional encryption
send "no\r"  ;# Send "no" followed by Enter to not use an existing wallet seed
send "OK\r"  ;# Send "OK" followed by Enter to proceed

expect eof

spawn ~/go/bin/btcwallet --simnet --noclienttls --noservertls -A "wallet/$filename" --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123 --rpcconnect 127.0.0.1:8000 --rpclisten 127.0.0.1:$wport
expect eof

interact