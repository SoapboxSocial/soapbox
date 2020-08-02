#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

setsebool httpd_can_network_connect on -P

sudo service nginx start

go run $GOPATH/src/github.com/ephemeral-networks/voicely/main.go &