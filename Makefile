all: peerviewer

peerviewer: *.go
	go get github.com/GeertJohan/go.rice
	go get github.com/GeertJohan/go.rice/rice
	go get github.com/Sirupsen/logrus
	go get github.com/gorilla/mux
	go get github.com/ziutek/glib
	go get github.com/ziutek/gst
	go get gopkg.in/alecthomas/kingpin.v2
	go get github.com/netCommonsEU/PeerStreamer-go-ml
	go get github.com/netCommonsEU/PeerStreamer-go-grapes
	go build -o peerviewer

# Requires sudo
packweb:
	cd public/ && npm install && webpack
	${GOPATH}/bin/rice append --exec peerviewer

install:
	mkdir -p /opt/peerstreamer
ifneq (,$(wildcard peerviewer))
	cp peerviewer /opt/peerstreamer/
	ln -f -s /opt/peerstreamer/peerviewer  /usr/local/bin/peerviewer
endif

uninstall:
ifneq (,$(wildcard /usr/local/bin/peerviewer))
	rm -f /usr/local/bin/peerviewer
	rm -f /opt/peerstreamer/peerviewer
endif
ifneq (,$(wildcard /opt/peerstreamer))
	rm -rf /opt/peerstreamer
endif

