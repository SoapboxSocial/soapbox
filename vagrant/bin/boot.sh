#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

setsebool httpd_can_network_connect on -P

sudo service nginx start
service postgresql start

sudo systemctl start supervisord
sudo systemctl enable supervisord
