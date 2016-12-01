PeerViewer
=================================================================

PeerViewer is a web-enabled front end for the [PeerStreamer project]
(https://github.com/netCommonsEU/PeerStreamer).

For checking out the current development code:

git clone -b D3.2-testing https://github.com/netCommonsEU/PeerStreamer-peerviewer

>### *PeerViewer on Raspberry Pi 2/3*
Building and installing PeerViewer on Raspberry Pi 2/3 requires a different
procedure than the one described below. The instructions for using PeerViewer on
Raspberry Pi 2/3 are reported [here](docs/raspberry_build.md).

## Requirements

Building PeerViewer requires the development versions of the following libraries:

* glib v2.0
* gstreamer v1.0
* libgstreamer-plugins-base1.0-dev

>### *NOTE*
PeerViewer requires GStreamer v1.8 or higher.

On ubuntu they can be installed with the following command:

```bash
sudo apt-get install libglib2.0-dev libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev gstreamer1.0-plugins-base gstreamer1.0-plugins-good gstreamer1.0-plugins-bad
```

Building PeerViewer also requires a properly configured Go development
environment. On Ubuntu (x86_64) it is possible to install and configure the go
development environment with the following commands:

```bash
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.7.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin # Put this in ~/.profile to make it permanent
mkdir -p ~/go_workspace
export GOPATH=~/go_workspace/ # Put this in ~/.profile to make it permanent 
```

For up-to-date instructions refer to the [official Go documentation]
(https://golang.org/doc/install)

Finally, also Node.js development environment is required. For configuring it on
Ubuntu (x84_64) execute the following commands:

```bash
curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
sudo apt-get install -y nodejs
sudo npm install webpack -g
```

Refer to the [official Node.js documentation]
(https://nodejs.org/en/download/package-manager/) for up-to-date instructions.

## Build

We recommend using the [PeerStreamer build system]
(https://github.com/netCommonsEU/PeerStreamer-build).

### Manual build

Just execute:

```bash
make
make packweb
```

this will produce an executable in the current directory named peerviewer.

## Install/Uninstall

For installing execute:

`make install`

For uninstalling execute:

`make uninstall`

## Basic Usage

Create a template configuration file and save it in config.json:

`./peerviewer --template > config.json`

config.json can be manually manipulated to suit your needs. Finally, start the
web server with the following command:

`./peerviewer -c config.json`

At this point you should be able to reach the web application pointing your
favorite browser at http://localhost:8080/

