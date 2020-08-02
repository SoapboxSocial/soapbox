#!/usr/bin/env bash

sudo yum install -y epel-release

sudo yum clean all

wget http://rpms.famillecollet.com/enterprise/remi-release-7.rpm
sudo rpm -Uvh remi-release-7*.rpm

sudo yum clean all

sudo yum install -y nginx

sudo yum install -y golang

echo 'export GOPATH="/home/vagrant/go"' >> ~/.bashrc
echo 'export PATH="$PATH:${GOPATH//://bin:}/bin"' >> ~/.bashrc
mkdir -p $GOPATH/{bin,pkg,src}

source ~/.bashrc

sudo rm -rf /etc/nginx/nginx.conf
sudo ln -s /vagrant/conf/nginx.conf /etc/nginx/nginx.conf

mkdir -p $GOPATH/src/github.com/ephemeral-networks/
sudo ln -s /var/www/ $GOPATH/src/github.com/ephemeral-networks/voicely

touch /vagrant/provisioned

echo "Provisioning done! Run 'vagrant reload'"