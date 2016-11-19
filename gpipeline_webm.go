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
	"strings"

	"github.com/ziutek/glib"
	"github.com/ziutek/gst"
)

type gPipelineVideoWebm struct {
	rtpSrcVideo      *gst.Element
	rtcpSrcVideo     *gst.Element
	rtpSrcAudio      *gst.Element
	rtcpSrcAudio     *gst.Element
	rtpbin	    *gst.Element
	vp8Depay    *gst.Element
	opusDepay   *gst.Element
	opusParser  *gst.Element
	oggMux	    *gst.Element
	sink        *gst.Element
	appSink     *gst.Element
	pipeline    *gst.Pipeline
	videoCaps   *gst.Caps
	audioCaps   *gst.Caps
	connections map[int]*net.TCPConn

	videoSessionID	uint32
	videoSessionConnected	bool
	audioSessionID	uint32
	audioSessionConnected	bool
}

func (p *gPipelineVideoWebm) needData(length uint32) {
	logger.Info("AppSrc needs data")
}

func (p *gPipelineVideoWebm) enoughData() {
	logger.Info("AppSrc has enough data")
}

func (p *gPipelineVideoWebm) newBuffer() {
	logger.Info("Received buffer")
}

func (p *gPipelineVideoWebm) rtpbinPadAdded(dec_sink_pad, demux_new_pad *gst.Pad) {
	var err bool
	logger.Info("Link filtered from rtpbin to vp8Depay")
	padname := C.gst_object_get_name((*C.GstObject)(demux_new_pad.GetPtr()))
	defer C.free(unsafe.Pointer(padname)) 
	videoprefix := fmt.Sprintf("recv_rtp_src_%d", p.videoSessionID)
	if strings.HasPrefix(C.GoString((*C.char)(padname)), videoprefix) && !p.videoSessionConnected {
		logger.Info("Video Prefix Found")
		err = p.rtpbin.LinkFiltered(p.vp8Depay, p.videoCaps)
		if !err {
			logger.Info("Link filtered from rtpbin to vp8Depay: FAILED")
			os.Exit(1)
		}
		p.videoSessionConnected = true

		return
	}

	audioprefix := fmt.Sprintf("recv_rtp_src_%d", p.audioSessionID)
	if strings.HasPrefix(C.GoString((*C.char)(padname)), audioprefix) && !p.audioSessionConnected {
		logger.Info("Audio Prefix Found")
		err = p.rtpbin.LinkFiltered(p.opusDepay, p.audioCaps)
		if !err {
			logger.Info("Link filtered from rtpbin to opusDepay: FAILED")
			os.Exit(1)
		}
		p.audioSessionConnected = true

		return
	}

}

func newGPipelineVideoWebm(id int) *gPipelineVideoWebm {
	var err bool
	p := gPipelineVideoWebm{}
	p.videoSessionID = 0
	p.audioSessionID = 1
	p.connections = make(map[int]*net.TCPConn)
	p.videoCaps = gst.NewCapsSimple("application/x-rtp", glib.Params{
		"media":         "video",
		"payload":       int32(96),
		"encoding-name": "VP8",
		"clock-rate": int32(90000),
	})
	p.audioCaps = gst.NewCapsSimple("application/x-rtp", glib.Params{
		"media":         "audio",
		"payload":       int32(96),
		"encoding-name": "OPUS",
		"clock-rate": int32(48000),
	})

	logger.Info("Create pipeline")
	pipe := gst.NewPipeline(fmt.Sprintf("stream-%d", id))
	p.pipeline = pipe

	logger.Info("Create RTP video Source (appsrc)")
	p.rtpSrcVideo = gst.ElementFactoryMake("appsrc", "RTP video Source")
	checkElem(p.rtpSrcVideo, "appsrc")
	p.rtpSrcVideo.SetProperty("is-live", true)
	//p.rtpSrcVideo.SetProperty("do-timestamp", true)
	p.rtpSrcVideo.SetProperty("format", uint32(3))
	sv := C.CString("caps")
        defer C.free(unsafe.Pointer(sv))
	C.g_object_set_go((*C.GObject)(p.rtpSrcVideo.GetPtr()), (*C.gchar)(sv), (unsafe.Pointer)(p.videoCaps))
	//p.rtpSrcVideo.ConnectNoi("need-data", p.needData, nil)
	//p.rtpSrcVideo.ConnectNoi("enough-data", p.enoughData, nil)

	logger.Info("Create RTCP video Source (appsrc)")
	p.rtcpSrcVideo = gst.ElementFactoryMake("appsrc", "RTCP video Source")
	checkElem(p.rtcpSrcVideo, "appsrc")
	p.rtcpSrcVideo.SetProperty("format", uint32(2))

	logger.Info("Create RTP audio Source (appsrc)")
	p.rtpSrcAudio = gst.ElementFactoryMake("appsrc", "RTP audio Source")
	checkElem(p.rtpSrcAudio, "appsrc")
	p.rtpSrcAudio.SetProperty("is-live", true)
	//p.rtpSrcVideo.SetProperty("do-timestamp", true)
	p.rtpSrcAudio.SetProperty("format", uint32(3))
	sa := C.CString("caps")
        defer C.free(unsafe.Pointer(sa))
	C.g_object_set_go((*C.GObject)(p.rtpSrcAudio.GetPtr()), (*C.gchar)(sa), (unsafe.Pointer)(p.audioCaps))
	//p.rtpSrcVideo.ConnectNoi("need-data", p.needData, nil)
	//p.rtpSrcVideo.ConnectNoi("enough-data", p.enoughData, nil)

	logger.Info("Create RTCP audio Source (appsrc)")
	p.rtcpSrcAudio = gst.ElementFactoryMake("appsrc", "RTCP audio Source")
	checkElem(p.rtcpSrcAudio, "appsrc")
	p.rtcpSrcAudio.SetProperty("format", uint32(2))

	logger.Info("Create rtpbin")
	p.rtpbin = gst.ElementFactoryMake("rtpbin", "rtpbin")
	checkElem(p.rtpbin, "rtpbin")

	logger.Info("Create rtpvp8depay")
	p.vp8Depay = gst.ElementFactoryMake("rtpvp8depay", "rtpvp8depay")
	checkElem(p.vp8Depay, "rtpvp8depay")

	logger.Info("Create rtpopusdepay")
	p.opusDepay = gst.ElementFactoryMake("rtpopusdepay", "rtpopusdepay")
	checkElem(p.opusDepay, "rtpopusdepay")

	logger.Info("Create opusparse")
	p.opusParser = gst.ElementFactoryMake("opusparse", "opusparse")
	checkElem(p.opusParser, "opusparse")

	logger.Info("Create muxer")
	//p.oggMux = gst.ElementFactoryMake("oggmux", "oggmux")
	//checkElem(p.oggMux, "oggmux")
	p.oggMux = gst.ElementFactoryMake("matroskamux", "oggmux")
	checkElem(p.oggMux, "matroskamux")
	p.oggMux.SetProperty("streamable", true)

	//logger.Info("Create opusparse")
	//p.opusParser = gst.ElementFactoryMake("opusparse", "opusparse")
	//checkElem(p.opusParser, "opusparse")

	logger.Info("Create sink (multifdsink)")
	p.sink = gst.ElementFactoryMake("multifdsink", "multifdsink")
	checkElem(p.sink, "multifdsink")
	p.sink.ConnectNoi("client-fd-removed", p.onClientFdRemoved, nil)
	//p.sink = gst.ElementFactoryMake("filesink", "filesink")
	//checkElem(p.sink, "filesink")
	//p.sink.SetProperty("location", "/tmp/prova.webm")

	logger.Info("Add elements to pipeline")
	pipe.Add(p.rtpSrcVideo, p.rtcpSrcVideo, p.rtpSrcAudio, p.rtcpSrcAudio, p.rtpbin, p.vp8Depay, p.opusDepay, p.opusParser, p.oggMux, p.sink)

	logger.Info("Get static pad for rtpsrc video")
	rtpSrcVideoPad := p.rtpSrcVideo.GetStaticPad("src")
	logger.Info("Get static pad for rtcpsrc video")
	rtcpSrcVideoPad := p.rtcpSrcVideo.GetStaticPad("src")

	logger.Info(fmt.Sprintf("Request pad recv_rtp_sink_%d", p.videoSessionID))
	rtpVideoSinkPad := p.rtpbin.GetRequestPad(fmt.Sprintf("recv_rtp_sink_%d", p.videoSessionID))
	logger.Info(fmt.Sprintf("Request pad recv_rtcp_sink_%d", p.videoSessionID))
	rtcpVideoSinkPad := p.rtpbin.GetRequestPad(fmt.Sprintf("recv_rtcp_sink_%d", p.videoSessionID))

	logger.Info("Link RTP src to RTP sink (video)")
	if !rtpSrcVideoPad.CanLink(rtpVideoSinkPad) {
		logger.Info("Link RTP src to RTP sink: FAILED")
		os.Exit(1)
	}
	rtpSrcVideoPad.Link(rtpVideoSinkPad)

	logger.Info("Link RTCP src to RTCP sink (video)")
	if !rtcpSrcVideoPad.CanLink(rtcpVideoSinkPad) {
		logger.Info("Link RTCP src to RTCP sink: FAILED")
		os.Exit(1)
	}
	rtcpSrcVideoPad.Link(rtcpVideoSinkPad)


	logger.Info("Get static pad for rtpsrc audio")
	rtpSrcAudioPad := p.rtpSrcAudio.GetStaticPad("src")
	logger.Info("Get static pad for rtcpsrc audio")
	rtcpSrcAudioPad := p.rtcpSrcAudio.GetStaticPad("src")

	logger.Info(fmt.Sprintf("Request pad recv_rtp_sink_%d", p.audioSessionID))
	rtpAudioSinkPad := p.rtpbin.GetRequestPad(fmt.Sprintf("recv_rtp_sink_%d", p.audioSessionID))
	logger.Info(fmt.Sprintf("Request pad recv_rtcp_sink_%d", p.audioSessionID))
	rtcpAudioSinkPad := p.rtpbin.GetRequestPad(fmt.Sprintf("recv_rtcp_sink_%d", p.audioSessionID))

	logger.Info("Link RTP src to RTP sink (audio)")
	if !rtpSrcAudioPad.CanLink(rtpAudioSinkPad) {
		logger.Info("Link RTP src to RTP sink: FAILED")
		os.Exit(1)
	}
	rtpSrcAudioPad.Link(rtpAudioSinkPad)

	logger.Info("Link RTCP src to RTCP sink (audio)")
	if !rtcpSrcAudioPad.CanLink(rtcpAudioSinkPad) {
		logger.Info("Link RTCP src to RTCP sink: FAILED")
		os.Exit(1)
	}
	rtcpSrcAudioPad.Link(rtcpAudioSinkPad)

	logger.Info("Link vp8Depay to oggMux")
	err = p.vp8Depay.Link(p.oggMux)
	if !err {
		logger.Info("Link vp8Depay to oggMux: FAILED")
		os.Exit(1)
	}

	logger.Info("Link opusDepay to opusParser to oggMux")
	err = p.opusDepay.Link(p.opusParser, p.oggMux)
	if !err {
		logger.Info("Link opusDepay to opusParser to oggMux: FAILED")
		os.Exit(1)
	}

	logger.Info("Link oggMux to sink")
	err = p.oggMux.Link(p.sink)
	if !err {
		logger.Info("Link oggMux to sink: FAILED")
		os.Exit(1)
	}

	logger.Info("Link filtered from rtpbin to vp8Depay")
	logger.Info("Register rtpbin pad callback")
	p.rtpbin.ConnectNoi("pad-added", p.rtpbinPadAdded, p.vp8Depay.GetStaticPad("sink"))
	p.rtpbin.ConnectNoi("pad-added", p.rtpbinPadAdded, p.opusDepay.GetStaticPad("sink"))

	logger.Info("Set state playing")
	state := pipe.SetState(gst.STATE_PLAYING)
	if state == gst.STATE_CHANGE_FAILURE {
		logger.Info("Set state playing: FAILED")
		os.Exit(0)
	}
	return &p
}

func (p *gPipelineVideoWebm) ListenForData(streams []rtpStream) {
	streamvideo := streams[1]
	streamaudio := streams[0]
	var buf []uint8
	for {
		select {
		case buf = <-streamvideo.RTP:
			logger.Info("Channel receive: streamvideo.RTP")
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtpSrcVideo.GetPtr()), test)
			//p.rtpSrcVideo.Emit("push-buffer", test)
			logger.Info("Buffer pushed (streamvideo.RTP)")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")
		case buf = <-streamvideo.RTCP:
			logger.Info("Channel receive: streamvideo.RTCP")
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtcpSrcVideo.GetPtr()), test)
			//p.rtcpSrcVideo.Emit("push-buffer", test)
			logger.Info("Buffer pushed (streamvideo.RTCP)")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")
		case buf = <-streamaudio.RTP:
			logger.Info("Channel receive: streamaudio.RTP")
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtpSrcAudio.GetPtr()), test)
			//p.rtpSrcVideo.Emit("push-buffer", test)
			logger.Info("Buffer pushed (streamaudio.RTP)")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")
		case buf = <-streamaudio.RTCP:
			logger.Info("Channel receive: streamaudio.RTCP")
			b := bytes.NewBuffer(buf)
			t := b.Bytes()
			logger.Info(len(t))
			test := C.gst_buffer_new_allocate(nil, C.gsize(len(t)), nil)
			C.gst_buffer_fill((unsafe.Pointer)(test), C.gsize(0), C.CBytes(t), C.gsize(len(t)))
			C.gst_app_src_push_buffer((unsafe.Pointer)(p.rtcpSrcAudio.GetPtr()), test)
			//p.rtcpSrcVideo.Emit("push-buffer", test)
			logger.Info("Buffer pushed (streamaudio.RTCP)")
			//C.gst_buffer_unref(test)
			//logger.Info("Buffer unref")

		}
	}
}

func (p *gPipelineVideoWebm) AddReceiver(conn *net.TCPConn) {
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

func (p *gPipelineVideoWebm) onClientFdRemoved(fd int32) {
	conn := p.connections[int(fd)]
	logger.Info("onClientFdRemoved")
	conn.Close()
	delete(p.connections, int(fd))
}

