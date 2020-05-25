# HW requirements: 2 CPUS 2 RAM
# OS: Ubuntu focal

#should be $(lsb_release -cs) but it appears it is not working on focal yet
docker_ubuntu_release=bionic
k8s_version=1.18.0

# Set hostname and disable swap
sudo hostnamectl set-hostname "$(hostname -I | awk '{print $1}')"
sudo swapoff -a && sudo sed -i '/ swap / s/^/#/' /etc/fstab

## Fix IP Addr
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys CC86BB64
sudo add-apt-repository ppa:rmescandon/yq -y
sudo apt-get update -y
sudo apt-get install yq -y
sudo yq w -i /etc/netplan/00-installer-config.yaml network.ethernets.enp0s3.dhcp4 false
sudo yq w -i /etc/netplan/00-installer-config.yaml network.ethernets.enp0s3.addresses[+] "$(hostname -I | awk '{print $1}')/24"
sudo yq w -i /etc/netplan/00-installer-config.yaml network.ethernets.enp0s3.gateway4 "$(hostname -I | cut -d "." -f 1-3).1"
sudo yq w -i /etc/netplan/00-installer-config.yaml network.ethernets.enp0s3.nameservers.addresses[+] 8.8.8.8
sudo yq w -i /etc/netplan/00-installer-config.yaml network.ethernets.enp0s3.nameservers.addresses[+] 8.8.4.4
sudo netplan apply

############# start k8s installation

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
  "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
  $docker_ubuntu_release \
  stable"
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
cat <<EOF | sudo tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF

sudo apt-get update
sudo apt-get install -y \
  containerd.io=1.2.13-1 \
  docker-ce=5:19.03.8~3-0~ubuntu-$docker_ubuntu_release \
  docker-ce-cli=5:19.03.8~3-0~ubuntu-$docker_ubuntu_release
sudo apt-get install -y kubelet=$k8s_version-00 kubeadm=$k8s_version-00 kubectl=$k8s_version-00
sudo apt-mark hold docker-ce kubelet kubeadm kubectl

echo "net.bridge.bridge-nf-call-iptables=1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

####### configure kubeadm

sudo kubeadm init --pod-network-cidr=10.244.0.0/16 &>~/kubeadm.log
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown "$(id -u):$(id -g)" $HOME/.kube/config
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/bc79dd1505b0c8681ece4de4c0d86c5cd2643275/Documentation/kube-flannel.yml
