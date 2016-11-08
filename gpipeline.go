package main

import "net"

type gPipeline interface {
	ListenForData(streams []rtpStream)
	AddReceiver(conn *net.TCPConn)
}
