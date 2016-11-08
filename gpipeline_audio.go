package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"syscall"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type gPipelineAudioOpus struct {
	rtpSrc      *gst.Element
	rtcpSrc     *gst.Element
	sink        *gst.Element
	pipeline    *gst.Pipeline
	connections map[int]*net.TCPConn
}

func newGPipelineAudioOpus(id int) *gPipelineAudioOpus {
	p := gPipelineAudioOpus{}
	p.connections = make(map[int]*net.TCPConn)
	p.rtpSrc = gst.ElementFactoryMake("appsrc", "RTP Source")
	audioCaps := gst.NewCapsSimple("application/x-rtp", glib.Params{
		"media":         "audio",
		"payload":       96,
		"encoding-name": "OPUS",
	})
	p.rtpSrc.SetProperty("is-live", true)

	p.rtcpSrc = gst.ElementFactoryMake("appsrc", "RTCP Source")

	rtpSrcPad := p.rtpSrc.GetStaticPad("src")
	rtcpSrcPad := p.rtcpSrc.GetStaticPad("src")

	rtpbin := gst.ElementFactoryMake("rtpbin", "rtpbin")
	rtpSinkPad := rtpbin.GetRequestPad("recv_rtp_sink_0")
	rtcpSinkPad := rtpbin.GetRequestPad("recv_rtcp_sink_0")

	rtpSrcPad.Link(rtpSinkPad)
	rtcpSrcPad.Link(rtcpSinkPad)

	opusDepay := gst.ElementFactoryMake("rtpopusdepay", "rtpopusdepay")
	oggMux := gst.ElementFactoryMake("oggmux", "oggmux")
	//opusDecoder := gst.ElementFactoryMake("opusdec", "opusdec")
	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)

	pipe := gst.NewPipeline(fmt.Sprintf("stream-%d", id))
	pipe.Add(p.rtpSrc, p.rtcpSrc, rtpbin, opusDepay, oggMux, p.sink)
	p.pipeline = pipe

	rtpbin.LinkFiltered(opusDepay, audioCaps)
	opusDepay.Link(oggMux, p.sink)

	pipe.SetState(gst.STATE_PLAYING)
	return &p
}

func (p *gPipelineAudioOpus) ListenForData(streams []rtpStream) {
	stream := streams[0]
	var buf []byte
	for {
		select {
		case buf = <-stream.RTP:
			p.rtpSrc.Emit("push-buffer", buf)
		case buf = <-stream.RTCP:
			p.rtcpSrc.Emit("push-buffer", buf)
		}
	}
}

func (p *gPipelineAudioOpus) AddReceiver(conn *net.TCPConn) {
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
		file, "HTTP/1.1 200 OK\r\nContent-Type: audio/opus\r\n\r\n",
	)
	if err != nil {
		log.Println("io.WriteString:", err)
		return
	}
	file.Close()

	p.connections[fd] = conn
	p.sink.Emit("add", fd)
}

func (p *gPipelineAudioOpus) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	conn.Close()
	delete(p.connections, int(fd))
}

type gPipelineAudioTest1 struct {
	rtpSrc      *gst.Element
	rtcpSrc     *gst.Element
	sink        *gst.Element
	pipeline    *gst.Pipeline
	connections map[int]*net.TCPConn
}

func newGPipelineAudioTest1(id int) *gPipelineAudioTest1 {
	p := gPipelineAudioTest1{}
	p.connections = make(map[int]*net.TCPConn)

	audioTestSrc := gst.ElementFactoryMake("audiotestsrc", "audiotestsrc")
	opusEncoder := gst.ElementFactoryMake("opusenc", "opusenc")
	oggMuxer := gst.ElementFactoryMake("oggmux", "oggmux")

	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)

	pipe := gst.NewPipeline(fmt.Sprintf("stream-%d", id))
	pipe.Add(audioTestSrc, opusEncoder, oggMuxer, p.sink)
	p.pipeline = pipe

	audioTestSrc.Link(opusEncoder, oggMuxer, p.sink)

	pipe.SetState(gst.STATE_PLAYING)
	return &p
}

func (p *gPipelineAudioTest1) ListenForData(streams []rtpStream) {}

func (p *gPipelineAudioTest1) AddReceiver(conn *net.TCPConn) {
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
		file, "HTTP/1.1 200 OK\r\nContent-Type: audio/opus\r\n\r\n",
	)
	if err != nil {
		log.Println("io.WriteString:", err)
		return
	}
	file.Close()

	p.connections[fd] = conn
	p.sink.Emit("add", fd)
}

func (p *gPipelineAudioTest1) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	conn.Close()
	delete(p.connections, int(fd))
}
