package main

/*
#cgo pkg-config: gstreamer-1.0
#cgo LDFLAGS: -lgstapp-1.0
#include <stdlib.h>
#include <string.h>
#include <gst/gst.h>
#include <gst/app/gstappsrc.h>

static void g_object_set_go(gpointer object, const gchar *first_property_name, void *val) {
	g_object_set(object, first_property_name, val, NULL);
}
*/
import "C"

import (
	"unsafe"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"bytes"
	"os"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type gPipelineVideoVP8 struct {
	rtpSrc      *gst.Element
	rtcpSrc     *gst.Element
	rtpbin	    *gst.Element
	vp8Depay   *gst.Element
	oggMux	    *gst.Element
	sink        *gst.Element
	appSink     *gst.Element
	pipeline    *gst.Pipeline
	audioCaps   *gst.Caps
	connections map[int]*net.TCPConn
}

func (p *gPipelineVideoVP8) needData(length uint32) {
	logger.Info("AppSrc needs data")
}

func (p *gPipelineVideoVP8) enoughData() {
	logger.Info("AppSrc has enough data")
}

func (p *gPipelineVideoVP8) newBuffer() {
	logger.Info("Received buffer")
}

func (p *gPipelineVideoVP8) rtpbinPadAdded(dec_sink_pad, demux_new_pad *gst.Pad) {
	var err bool
	logger.Info("Link filtered from rtpbin to vp8Depay")
	err = p.rtpbin.LinkFiltered(p.vp8Depay, p.audioCaps)
	if !err {
		logger.Info("Link filtered from rtpbin to vp8Depay: FAILED")
		os.Exit(1)
	}
}

func newGPipelineVideoVP8(id int) *gPipelineVideoVP8 {
	var err bool
	p := gPipelineVideoVP8{}
	p.connections = make(map[int]*net.TCPConn)
	p.audioCaps = gst.NewCapsSimple("application/x-rtp", glib.Params{
		"media":         "video",
		"payload":       int32(96),
		"encoding-name": "VP8",
		"clock-rate": int32(90000),
	})

	logger.Info("Create pipeline")
	pipe := gst.NewPipeline(fmt.Sprintf("stream-%d", id))
	p.pipeline = pipe

	logger.Info("Create RTP Source (appsrc)")
	p.rtpSrc = gst.ElementFactoryMake("appsrc", "RTP Source")
	checkElem(p.rtpSrc, "appsrc")
	p.rtpSrc.SetProperty("is-live", true)
	//p.rtpSrc.SetProperty("do-timestamp", true)
	p.rtpSrc.SetProperty("format", uint32(3))
	s := C.CString("caps")
        defer C.free(unsafe.Pointer(s))
	C.g_object_set_go((*C.GObject)(p.rtpSrc.GetPtr()), (*C.gchar)(s), (unsafe.Pointer)(p.audioCaps))
	p.rtpSrc.ConnectNoi("need-data", p.needData, nil)
	p.rtpSrc.ConnectNoi("enough-data", p.enoughData, nil)

	logger.Info("Create RTCP Source (appsrc)")
	p.rtcpSrc = gst.ElementFactoryMake("appsrc", "RTCP Source")
	checkElem(p.rtcpSrc, "appsrc")
	p.rtcpSrc.SetProperty("format", uint32(2))

	logger.Info("Create rtpbin")
	p.rtpbin = gst.ElementFactoryMake("rtpbin", "rtpbin")
	checkElem(p.rtpbin, "rtpbin")

	logger.Info("Create rtpvp8depay")
	p.vp8Depay = gst.ElementFactoryMake("rtpvp8depay", "rtpvp8depay")
	checkElem(p.vp8Depay, "rtpvp8depay")

	logger.Info("Create muxer")
	//p.oggMux = gst.ElementFactoryMake("oggmux", "oggmux")
	//checkElem(p.oggMux, "oggmux")
	p.oggMux = gst.ElementFactoryMake("matroskamux", "oggmux")
	checkElem(p.oggMux, "matroskamux")
	p.oggMux.SetProperty("streamable", true)

	logger.Info("Create sink (multifdsink)")
	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	checkElem(p.sink, "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)
	//p.sink = gst.ElementFactoryMake("filesink", "filesink")
	//checkElem(p.sink, "filesink")
	//p.sink.SetProperty("location", "/tmp/prova.webm")

	logger.Info("Add elements to pipeline")
	pipe.Add(p.rtpSrc, p.rtcpSrc, p.rtpbin, p.vp8Depay, p.oggMux, p.sink)

	logger.Info("Get static pad for rtpsrc")
	rtpSrcPad := p.rtpSrc.GetStaticPad("src")
	logger.Info("Get static pad for rtcpsrc")
	rtcpSrcPad := p.rtcpSrc.GetStaticPad("src")

	logger.Info("Request pad recv_rtp_sink")
	rtpSinkPad := p.rtpbin.GetRequestPad("recv_rtp_sink_0")
	logger.Info("Request pad recv_rtcp_sink")
	rtcpSinkPad := p.rtpbin.GetRequestPad("recv_rtcp_sink_0")

	logger.Info("Link RTP src to RTP sink")
	if !rtpSrcPad.CanLink(rtpSinkPad) {
		logger.Info("Link RTP src to RTP sink: FAILED")
		os.Exit(1)
	}
	rtpSrcPad.Link(rtpSinkPad)

	logger.Info("Link RTCP src to RTCP sink")
	if !rtcpSrcPad.CanLink(rtcpSinkPad) {
		logger.Info("Link RTCP src to RTCP sink: FAILED")
		os.Exit(1)
	}
	rtcpSrcPad.Link(rtcpSinkPad)



	logger.Info("Link vp8Depay to oggMux and oggMux to sink")
	err = p.vp8Depay.Link(p.oggMux, p.sink)
	if !err {
		logger.Info("Link vp8Depay to oggMux and oggMux to sink: FAILED")
		os.Exit(1)
	}

	logger.Info("Link filtered from rtpbin to vp8Depay")
	logger.Info("Register rtpbin pad callback")
	p.rtpbin.ConnectNoi("pad-added", p.rtpbinPadAdded, p.vp8Depay.GetStaticPad("sink"))

	logger.Info("Set state playing")
	state := pipe.SetState(gst.STATE_PLAYING)
	if state == gst.STATE_CHANGE_FAILURE {
		logger.Info("Set state playing: FAILED")
		os.Exit(0)
	}
	return &p
}

func (p *gPipelineVideoVP8) ListenForData(streams []rtpStream) {
	stream := streams[0]
	var buf []uint8
	for {
		select {
		case buf = <-stream.RTP:
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtpSrc.GetPtr()), test)
			//p.rtpSrc.Emit("push-buffer", test)
			logger.Info("Buffer pushed")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")
		case buf = <-stream.RTCP:
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtcpSrc.GetPtr()), test)
			//p.rtcpSrc.Emit("push-buffer", test)
			logger.Info("Buffer pushed")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")
		}
	}
}

func (p *gPipelineVideoVP8) AddReceiver(conn *net.TCPConn) {
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
		file, "HTTP/1.1 200 OK\r\nContent-Type: audio/webm\r\n\r\n",
	)
	if err != nil {
		log.Println("io.WriteString:", err)
		return
	}
	file.Close()

	logger.Info("Add receiver")
	p.connections[fd] = conn
	p.sink.Emit("add", fd)
}

func (p *gPipelineVideoVP8) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	logger.Info("onClientFdRemoved")
	conn.Close()
	delete(p.connections, int(fd))
}


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
