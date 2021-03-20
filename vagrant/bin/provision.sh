#!/usr/bin/env bash

sudo echo nameserver 8.8.8.8 >> /etc/resolv.conf
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

sudo yum install -y redis

rm -rf /etc/supervisord.conf
sudo ln -s /vagrant/conf/supervisord.conf /etc/supervisord.conf
sudo mkdir -p /etc/supervisor/conf.d/
sudo ln -s /vagrant/conf/supervisord/soapbox.conf /etc/supervisor/conf.d/soapbox.conf
sudo ln -s /vagrant/conf/supervisord/notifications.conf /etc/supervisor/conf.d/notifications.conf
sudo ln -s /vagrant/conf/supervisord/indexer.conf /etc/supervisor/conf.d/indexer.conf
sudo ln -s /vagrant/conf/supervisord/rooms.conf /etc/supervisor/conf.d/rooms.conf
sudo ln -s /vagrant/conf/supervisord/metadata.conf /etc/supervisor/conf.d/metadata.conf

echo 'export GOPATH="/home/vagrant/go"' >> ~/.bashrc
echo 'export PATH="$PATH:${GOPATH//://bin:}/bin"' >> ~/.bashrc
mkdir -p $GOPATH/{bin,pkg,src}

source ~/.bashrc

sudo rpm -Uvh https://download.postgresql.org/pub/repos/yum/reporpms/EL-8-x86_64/pgdg-redhat-repo-latest.noarch.rpm
sudo yum install -y postgresql96-server postgresql96
sudo /usr/pgsql-9.6/bin/postgresql96-setup initdb

sudo systemctl start postgresql-9.6
sudo systemctl enable postgresql-9.6

sudo su - postgres -c "psql -a -w -f /var/www/db/database.sql"
sudo su - postgres -c "psql -t voicely -a -w -f /var/www/db/tables.sql"

#rm /var/lib/pgsql/9.6/data/postgresql.conf
#ln -s /vagrant/conf/postgres.conf /var/lib/pgsql/9.6/data/postgresql.conf

wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.8.1-x86_64.rpm
wget https://artifacts.elastic.co/downloads/elasticsearch/elasticsearch-7.8.1-x86_64.rpm.sha512
sudo rpm --install elasticsearch-7.8.1-x86_64.rpm

sudo rm -rf /etc/nginx/nginx.conf
sudo ln -s /vagrant/conf/nginx.conf /etc/nginx/nginx.conf

mkdir -p $GOPATH/src/github.com/soapboxsocial/
sudo ln -s /var/www/ $GOPATH/src/github.com/soapboxsocial/soapbox

mkdir -p /conf/services
sudo cp -p sudo cp -R /var/www/conf/services/* /conf/services
sudo chown nginx:nginx -R /conf/services

sudo ln -s $GOPATH/src/github.com/soapboxsocial/soapbox/conf/services/ /conf/services

sudo mkdir -p /cdn/images/groups/
sudo chown nginx:nginx -R /cdn/images
sudo chmod -R 0777 /cdn/images

sudo mkdir -p /cdn/stories/
sudo chown nginx:nginx -R /cdn/stories
sudo chmod -R 0777 /cdn/stories

cd $GOPATH/src/github.com/soapboxsocial/soapbox && sudo go build -o /usr/local/bin/soapbox main.go
cd $GOPATH/src/github.com/soapboxsocial/soapbox/cmd/indexer && sudo go build -o /usr/local/bin/indexer main.go
cd $GOPATH/src/github.com/soapboxsocial/soapbox/cmd/rooms && sudo go build -o /usr/local/bin/rooms main.go
cd $GOPATH/src/github.com/soapboxsocial/soapbox/cmd/stories && sudo go build -o /usr/local/bin/stories main.go

crontab /vagrant/conf/crontab

touch /vagrant/provisioned

echo "Provisioning done! Run 'vagrant reload'"
