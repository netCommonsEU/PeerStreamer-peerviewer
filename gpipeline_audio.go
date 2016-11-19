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

type gPipelineAudioOpus struct {
	rtpSrc      *gst.Element
	rtcpSrc     *gst.Element
	rtpbin	    *gst.Element
	opusDepay   *gst.Element
	oggMux	    *gst.Element
	opusParser  *gst.Element
	sink        *gst.Element
	appSink     *gst.Element
	pipeline    *gst.Pipeline
	audioCaps   *gst.Caps
	connections map[int]*net.TCPConn
}

func checkElem(e *gst.Element, name string) {
        if e == nil {
                fmt.Fprintln(os.Stderr, "can't make element: ", name)
                os.Exit(1)
        }
}

func (p *gPipelineAudioOpus) needData(length uint32) {
	logger.Info("AppSrc needs data")
}

func (p *gPipelineAudioOpus) enoughData() {
	logger.Info("AppSrc has enough data")
}

func (p *gPipelineAudioOpus) newBuffer() {
	logger.Info("Received buffer")
}

func (p *gPipelineAudioOpus) rtpbinPadAdded(dec_sink_pad, demux_new_pad *gst.Pad) {
	var err bool
	logger.Info("Link filtered from rtpbin to opusDepay")
	err = p.rtpbin.LinkFiltered(p.opusDepay, p.audioCaps)
	if !err {
		logger.Info("Link filtered from rtpbin to opusDepay: FAILED")
		os.Exit(1)
	}
}

func newGPipelineAudioOpus(id int) *gPipelineAudioOpus {
	var err bool
	p := gPipelineAudioOpus{}
	p.connections = make(map[int]*net.TCPConn)
	p.audioCaps = gst.NewCapsSimple("application/x-rtp", glib.Params{
		"media":         "audio",
		"payload":       int32(96),
		"encoding-name": "OPUS",
		"clock-rate": int32(48000),
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
	p.rtcpSrc.SetProperty("format", uint32(3))

	logger.Info("Create rtpbin")
	p.rtpbin = gst.ElementFactoryMake("rtpbin", "rtpbin")
	checkElem(p.rtpbin, "rtpbin")

	logger.Info("Create rtpopusdepay")
	p.opusDepay = gst.ElementFactoryMake("rtpopusdepay", "rtpopusdepay")
	checkElem(p.opusDepay, "rtpopusdepay")

	logger.Info("Create muxer")
	p.oggMux = gst.ElementFactoryMake("oggmux", "oggmux")
	checkElem(p.oggMux, "oggmux")
	//p.oggMux = gst.ElementFactoryMake("matroskamux", "oggmux")
	//checkElem(p.oggMux, "matroskamux")

	logger.Info("Create opusparse")
	p.opusParser = gst.ElementFactoryMake("opusparse", "opusparse")
	checkElem(p.opusParser, "opusparse")

	logger.Info("Create sink (multifdsink)")
	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	checkElem(p.sink, "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)
	//p.sink = gst.ElementFactoryMake("filesink", "filesink")
	//checkElem(p.sink, "filesink")
	//p.sink.SetProperty("location", "/tmp/prova.webm")

	logger.Info("Add elements to pipeline")
	pipe.Add(p.rtpSrc, p.rtcpSrc, p.rtpbin, p.opusDepay, p.opusParser, p.oggMux, p.sink)

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



	logger.Info("Link opusDepay to oggMux and oggMux to sink")
	err = p.opusDepay.Link(p.opusParser, p.oggMux, p.sink)
	if !err {
		logger.Info("Link opusDepay to oggMux and oggMux to sink: FAILED")
		os.Exit(1)
	}

	logger.Info("Link filtered from rtpbin to opusDepay")
	logger.Info("Register rtpbin pad callback")
	p.rtpbin.ConnectNoi("pad-added", p.rtpbinPadAdded, p.opusDepay.GetStaticPad("sink"))

	logger.Info("Set state playing")
	state := pipe.SetState(gst.STATE_PLAYING)
	if state == gst.STATE_CHANGE_FAILURE {
		logger.Info("Set state playing: FAILED")
		os.Exit(0)
	}
	return &p
}

func (p *gPipelineAudioOpus) ListenForData(streams []rtpStream) {
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

	logger.Info("Add receiver")
	p.connections[fd] = conn
	p.sink.Emit("add", fd)
}

func (p *gPipelineAudioOpus) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	logger.Info("onClientFdRemoved")
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
