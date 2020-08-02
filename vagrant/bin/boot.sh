#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

setsebool httpd_can_network_connect on -P

sudo service nginx start
service postgresql start

go run $GOPATH/src/github.com/ephemeral-networks/voicely/main.go > /var/log/voicely.log 2>&1 &