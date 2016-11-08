PeerViewer
=================================================================

PeerViewer is a web-enabled front end for the [PeerStreamer project]
(https://github.com/netCommonsEU/PeerStreamer).


## Requirements

Building PeerViewer requires the developers versions of the following libraries:

* glib v2.0
* gstreamer v1.0

On ubuntu they can be installed with the following command:

`sudo apt-get install libglib2.0-dev libgstreamer1.0-dev`

Building PeerViewer also requires a properly configured Go development
environment. On Ubuntu (x86_64) it is possible to install and configure the go
development environment with the following commands:

```
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.7.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin # Put this in ~/.profile to make it permanent
mkdir -p ~/go_workspace
export GOPATH=~/go_workspace/ # Put this in ~/.profile to make it permanent 
```

For up-to-date instructions refer to the [official Go documentation]
(https://golang.org/doc/install)

## Build

Just execute:

`make`

this will produce an executable in the current directory named peerviewer.

