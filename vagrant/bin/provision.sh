#!/usr/bin/env bash

sudo yum install -y epel-release

sudo yum clean all

sudo yum remove git*
sudo yum -y install https://packages.endpoint.com/rhel/7/os/x86_64/endpoint-repo-1.7-1.x86_64.rpm
sudo yum -y install git

wget http://rpms.famillecollet.com/enterprise/remi-release-7.rpm
sudo rpm -Uvh remi-release-7*.rpm

sudo yum clean all

sudo yum install -y nginx

sudo yum install -y golang

sudo yum install -y supervisor

sudo yum install redis

rm -rf /etc/supervisord.conf
sudo ln -s /vagrant/conf/supervisord.conf /etc/supervisord.conf
sudo mkdir -p /etc/supervisor/conf.d/
sudo ln -s /vagrant/conf/voicely.conf /etc/supervisor/conf.d/voicely.conf

echo 'export GOPATH="/home/vagrant/go"' >> ~/.bashrc
echo 'export PATH="$PATH:${GOPATH//://bin:}/bin"' >> ~/.bashrc
mkdir -p $GOPATH/{bin,pkg,src}

source ~/.bashrc

yum install -y postgresql-server postgresql-contrib
postgresql-setup initdb

systemctl start postgresql
systemctl enable postgresql

sudo su - postgres -c "psql -a -w -f /var/www/db/database.sql"
sudo su - postgres -c "psql -t voicely -a -w -f /var/www/db/tables.sql"

rm /var/lib/pgsql/data/pg_hba.conf
ln -s /vagrant/conf/postgres.conf /var/lib/pgsql/data/pg_hba.conf

sudo rm -rf /etc/nginx/nginx.conf
sudo ln -s /vagrant/conf/nginx.conf /etc/nginx/nginx.conf

mkdir -p $GOPATH/src/github.com/ephemeral-networks/
sudo ln -s /var/www/ $GOPATH/src/github.com/ephemeral-networks/voicely

cd $GOPATH/src/github.com/ephemeral-networks/voicely && sudo go build -o /usr/local/bin/voicely main.go

touch /vagrant/provisioned

echo "Provisioning done! Run 'vagrant reload'"