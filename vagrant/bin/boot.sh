#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

setsebool httpd_can_network_connect on -P

sudo service nginx start
sudo service postgresql start
sudo service redis start

sudo systemctl start supervisord
sudo systemctl enable supervisord
