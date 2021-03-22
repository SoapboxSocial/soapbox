#!/usr/bin/env bash

PATH=$PATH:/usr/local/bin

setsebool httpd_can_network_connect on -P

sudo service nginx start
sudo service postgresql-9.6 start
sudo service redis start
sudo service elasticsearch start

sudo systemctl start supervisord
sudo systemctl enable supervisord
