package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"syscall"

	"github.com/ziutek/gst"
)

type gPipelineVideoTest1 struct {
	rtpSrc      *gst.Element
	rtcpSrc     *gst.Element
	sink        *gst.Element
	pipeline    *gst.Pipeline
	connections map[int]*net.TCPConn
}

func newGPipelineVideoTest1(id int) *gPipelineVideoTest1 {
	p := gPipelineVideoTest1{}
	p.connections = make(map[int]*net.TCPConn)

	videoTestSrc := gst.ElementFactoryMake("videotestsrc", "videotestsrc")
	vp8Encoder := gst.ElementFactoryMake("vp8enc", "vp8enc")
	webmMuxer := gst.ElementFactoryMake("webmmux", "webmmux")

	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)

	pipe := gst.NewPipeline(fmt.Sprintf("stream-%d", id))
	pipe.Add(videoTestSrc, vp8Encoder, webmMuxer, p.sink)
	p.pipeline = pipe

	videoTestSrc.Link(vp8Encoder, webmMuxer, p.sink)

	pipe.SetState(gst.STATE_PLAYING)
	return &p
}

func (p *gPipelineVideoTest1) ListenForData(streams []rtpStream) {}

func (p *gPipelineVideoTest1) AddReceiver(conn *net.TCPConn) {
	file, err := conn.File()
	if err != nil {
		log.Println("net.TCPConn.File:", err)
		return
	}
	fd, err := syscall.Dup(int(file.Fd()))
	if err != nil {
		log.Println("syscall.Dup:", err)
		return
	}
	// Send HTTP header
	_, err = io.WriteString(
		file, "HTTP/1.1 200 OK\r\nContent-Type: video/webm\r\n\r\n",
	)
	if err != nil {
		log.Println("io.WriteString:", err)
		return
	}
	file.Close()

	p.connections[fd] = conn
	p.sink.Emit("add", fd)
}

func (p *gPipelineVideoTest1) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	conn.Close()
	delete(p.connections, int(fd))
}
