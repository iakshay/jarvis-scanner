Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.provision "shell", inline: "echo Hello"
  config.vm.memory = 512
  config.vm.cpus = 1
  
	config.vm.define "client" do |client|
    client.vm.hostname = "client"
    client.vm.network :private_network, ip: "10.0.0.10"

    # Install NFS client
    client.vm.provision "shell", inline: <<-SHELL
        apt-get update
        add-apt-repository ppa:longsleep/golang-backports
        apt-get -y install golang-go
    SHELL
  end

  config.vm.define "server" do |server|
    server.vm.hostname = "server"
    server.vm.network :private_network, ip: "10.0.0.11"

    # Install NFS server
    server.vm.provision "shell", inline: <<-SHELL
        apt-get update
        add-apt-repository ppa:longsleep/golang-backports
        apt-get -y install golang-go
    SHELL
  end
end

