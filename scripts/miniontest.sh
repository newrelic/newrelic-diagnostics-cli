#!/usr/bin/expect -f
set arg [lrange $argv 0 0]

#This is the IP of your private minion
set server "${arg}"

set name "synthetics"
set pass "synthetics"
set filename "nrdiag"

#Build binary for linux
spawn env GOOS=linux GOARCH=386 go build -o $filename -v
expect "$ "

#Move binary to /bin/linux
spawn mv "$filename" "bin/linux/$filename"
expect "$ "

#Upload binary to minion
spawn sftp $name@$server
expect "*?assword:*"
send -- "$pass\r"
expect "sftp>"
send -- "put bin/linux/$filename\r"
expect "sftp>"
send "quit\r"

#SSH into minion
spawn ssh $name@$server -oStrictHostKeyChecking=no -oUserKnownHostsFile=/dev/null
match_max 100000
expect "*?assword:*"
send -- "$pass\r"
expect "$ "
send -- "sudo su\r"
expect "*sudo*"
send -- "$pass\r"
expect "# "
send -- "mv $filename /opt/newrelic/synthetics/$filename\r"
expect "# "
send -- "cd /opt/newrelic/synthetics\r"

#Execute binary
send -- "./$filename -t Synthetics/*\r"
interact