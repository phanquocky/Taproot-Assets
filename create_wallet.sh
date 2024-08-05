#!/usr/bin/expect -f

set password "admin"

spawn ~/go/bin/btcwallet --create --simnet --noclienttls --noservertls -A wallet --btcdusername=admin --btcdpassword=admin123 -u admin -P admin123

expect "Enter the private passphrase for your new wallet:" 
send "$password\r"

expect "Confirm passphrase:"
send "$password\r"

send "no\r"  ;# Send "no" followed by Enter to opt out of additional encryption
send "no\r"  ;# Send "no" followed by Enter to not use an existing wallet seed
send "OK\r"  ;# Send "OK" followed by Enter to proceed

expect eof
