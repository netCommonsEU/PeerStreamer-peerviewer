all: peerviewer

peerviewer:
	go get github.com/GeertJohan/go.rice
	go get github.com/Sirupsen/logrus
	go get github.com/gorilla/mux
	go get github.com/ziutek/glib
	go get github.com/ziutek/gst
	go get gopkg.in/alecthomas/kingpin.v2
	go get github.com/netCommonsEU/PeerStreamer-go-ml
	go get github.com/netCommonsEU/PeerStreamer-go-grapes
	go build
	mv PeerStreamer-peerviewer peerviewer
