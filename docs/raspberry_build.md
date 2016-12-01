PeerViewer on Raspberry Pi 2/3
=================================================================

Below are reported the step-by-step instructions for building and installing
PeerViewer on a Raspberry Pi 2/3 device running Raspbian Jessie Lite. Currently,
the procedure has been tested on Raspbian Jessie Lite Version: September 2016.
Given that PeerViewer requires GStreamer v1.8 or higher, but Raspbian Jessie
Lite is shipped by default with GStreamer v1.8 we will need to retrieve the
required packages from the Stretch repositories. This is an experimental
procedure, don't do this on a production node.


## Requirements

Add the Stretch repository to the repositories source list:

```bash
echo 'deb http://mirrordirector.raspbian.org/raspbian/ stretch main contrib non-free rpi' | sudo tee -a /etc/apt/sources.list
```

Update the repositories cache and upgrade all the packages (be aware that this
can take a long time):

```bash
sudo apt-get update
sudo apt-get --with-new-pkgs upgrade
```

Install the PeerViewer specific requirements:

```bash
sudo apt-get install libglib2.0-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev libgstreamer-plugins-bad1.0-dev gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad git zip
```

### Set up the Go development environment

This is not different that the standard procedure for installing Go:

```bash
wget https://storage.googleapis.com/golang/go1.7.3.linux-armv6l.tar.gz
sudo tar -C /usr/local -xzf go1.7.3.linux-armv6l.tar.gz
export PATH=$PATH:/usr/local/go/bin # Put this in ~/.profile to make it permanent
mkdir -p ~/go_workspace
export GOPATH=~/go_workspace/ # Put this in ~/.profile to make it permanent
```

For up-to-date instructions refer to the [official Godocumentation] (https://golang.org/doc/install)


### Set up the Node.js development environment

Install Node.js and webpack:

```bash
curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install webpack -g
```

Install PhantomJS for ARM architecture:

```bash
sudo apt-get install libfontconfig1 libfreetype6 libpng12-0
curl -o /tmp/phantomjs -sSL
https://github.com/fg2it/phantomjs-on-raspberry/releases/download/v2.1.1-wheezy-jessie/phantomjs
sudo mv /tmp/phantomjs /usr/local/bin/phantomjs
sudo chmod a+x /usr/local/bin/phantomjs
```

## PeerViewer Build and Install

For building and installing PeerViewer execute the following commands:

```bash
git clone -b D3.2-testing --depth 1 https://github.com/netCommonsEU/PeerStreamer-peerviewer
cd PeerStreamer-peerviewer
make
make packweb
sudo make install
```

## Testing

For basic PeerViewer tests see the [testing instructions]
(https://github.com/netCommonsEU/PeerStreamer/blob/D3.2-testing/testing/Testing.md)

