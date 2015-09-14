# -*- mode: ruby -*-
# vi: set ft=ruby :

ENV['VAGRANT_DEFAULT_PROVIDER'] = "virtualbox"

Vagrant.configure(2) do |config|
  config.vm.box = "phusion/ubuntu-14.04-amd64"
  config.vm.synced_folder "./", "/home/vagrant/cilium"
  config.vm.provision "shell", inline: <<-SHELL
  sudo apt-get update
  sudo su -c 'wget -qO- https://get.docker.com/ | sh'
  sudo dpkg -i /home/vagrant/cilium/vagrant/openvswitch-switch_2.4.0-1_amd64.deb /home/vagrant/cilium/vagrant/openvswitch-common_2.4.0-1_amd64.deb
  sudo cp /home/vagrant/cilium/vagrant/dockerconfig /etc/default/docker
  sudo service docker restart
  sleep 5s
  sudo usermod -aG docker vagrant
  /home/vagrant/cilium/scripts/import-images.sh
  SHELL

  config.vm.define "node1" do |node1|
    node1.vm.network "private_network", ip: "192.168.50.5"
    node1.vm.hostname = "node1"
  end

  config.vm.define "node2" do |node2|
    node2.vm.network "private_network", ip: "192.168.50.6"
    node2.vm.hostname = "node2"
  end

end
