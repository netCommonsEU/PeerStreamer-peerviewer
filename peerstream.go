package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/netCommonsEU/PeerStreamer-go-grapes"
	"github.com/netCommonsEU/PeerStreamer-go-ml"
)

type rtpStream struct {
	RTP  chan []uint8
	RTCP chan []uint8
}

type psTemporalBuffer struct {
	PacketAssembly *ml.PacketAssembly
	LastUpdated    time.Time
}

type pStream struct {
	inboundConn      *net.UDPConn
	assemblyBuffers  map[uint32]*psTemporalBuffer
	assemblyLifetime time.Duration
	rtpStreams       []rtpStream
}

func newPStream(conn *net.UDPConn, numMediaStreams int) *pStream {
	stream := pStream{}
	stream.inboundConn = conn
	stream.assemblyLifetime = 3 * time.Second
	stream.assemblyBuffers = make(map[uint32]*psTemporalBuffer)
	stream.rtpStreams = make([]rtpStream, numMediaStreams)
	for i := 0; i < numMediaStreams; i++ {
		stream.rtpStreams[i].RTP = make(chan []byte)
		stream.rtpStreams[i].RTCP = make(chan []byte)
	}
	return &stream
}

func (s *pStream) RTPStreams() []rtpStream {
	return s.rtpStreams[:]
}

func (s *pStream) ListenInbound() {
	for {
		data := make([]byte, 65536)
		n, err := s.inboundConn.Read(data)
		if err != nil {
			logger.Error("Cannot read() on UDP socket: ", err)
			continue // maybe break/return
		}
		packet, err := ml.ParsePacket(data[:n])
		if err != nil {
			logger.Warning("Cannot parse UDP packet: ", err)
			continue
		}
		// easy case: the packet doesn't need to be re-assembled
		if packet.ContentOffset == 0 && len(packet.Content) == int(packet.ContentTotalSize) {
			logger.WithField("seq", packet.Sequence).Debug("Dispatching UDP packet content to GRAPES")
			go s.handleGrapesMessage(packet.Content)
			continue
		}
		// re-assembly needed
		buf, ok := s.assemblyBuffers[packet.Sequence]
		logFields := log.Fields{
			"seq":    packet.Sequence,
			"offset": packet.ContentOffset,
			"total":  packet.ContentTotalSize,
		}
		if !ok {
			// first packet for this sequence or sequence too old
			logger.WithFields(logFields).Debug("Received fragment of a new packet")
			buf = &psTemporalBuffer{PacketAssembly: ml.NewPacketAssembly(packet), LastUpdated: time.Now()}
			s.assemblyBuffers[packet.Sequence] = buf
		}
		logger.WithFields(logFields).Debug("Adding packet to assembly")
		buf.LastUpdated = time.Now()
		buf.PacketAssembly.Push(packet)
		if buf.PacketAssembly.Ready() {
			go s.handleGrapesMessage(buf.PacketAssembly.Buffer)
			logger.WithFields(logFields).Debug("Dispatching UDP packet content to GRAPES")
			delete(s.assemblyBuffers, packet.Sequence)
		}
	}
}

func (s *pStream) handleGrapesMessage(data []byte) {
	grapesMsg, err := grapes.ParseMessage(data)
	if err != nil {
		logger.Warning("Cannot parse GRAPES message: ", err)
		return
	}
	switch grapesMsg.Type {
	case grapes.TypeChunk:
		logger.WithField("transaction", grapesMsg.TransactionID).Debug("Message contains chunks, processing")
		s.handleChunks(grapesMsg)
	default:
		// ignore
		logger.WithField("transaction", grapesMsg.TransactionID).Debug("Message doesn't contains chunks, ignoring")
	}
}

func (s *pStream) handleChunks(msg *grapes.Message) {
	l := len(msg.Content)
	for consumed := 0; consumed < l; {
		chunk, b, err := grapes.ParseChunk(msg.Content[consumed:])
		if err != nil {
			logger.WithFields(log.Fields{
				"transaction": msg.TransactionID,
				"chunk":       chunk.ID,
				"timestamp":   chunk.Timestamp,
				"offset":      consumed,
			}).Warning("Cannot parse GRAPES chunk: ", err)
			return
		}
		logger.WithFields(log.Fields{
			"transaction": msg.TransactionID,
			"offset":      consumed,
			"chunk":       chunk.ID,
			"timestamp":   chunk.Timestamp,
		}).Debug("Dispatching RTP envelopes")
		s.handleRTPEnvelopes(chunk)
		consumed += int(b)
	}
}

func (s *pStream) handleRTPEnvelopes(chunk *grapes.Chunk) {
	l := len(chunk.Content)
	for consumed := 0; consumed < l; {
		e, b, err := grapes.ParseRTPEnvelope(chunk.Content[consumed:])
		if err != nil {
			logger.WithFields(log.Fields{
				"chunk":  chunk.ID,
				"offset": consumed,
				"stream": e.StreamID,
			}).Warn("Cannot parse RTP envelope: ", err)
			return
		}
		logger.WithFields(log.Fields{
			"chunk":  chunk.ID,
			"offset": consumed,
			"stream": e.StreamID,
		}).Debug("Dispatching RTP/RTCP packet")
		s.dispatchRTPPackets(e)
		consumed += int(b)
	}
}

func (s *pStream) dispatchRTPPackets(env *grapes.RTPEnvelope) {
	/*
	 * WARNING: Streams management still incomplete.
	 * For now we assume two scenarios:
	 * 1) sigle stream: only audio (OPUS) or only video (VP8)
	 * 2) video + audio streams: video identified by env.StreamID == 2
	 * and audio identified by env.StreamID == 1 (We assume the RTP chunkiser works in this way).
	 * we push audio RTP/RTCP packets in s.rtpStreams[0] and video
	 * RTP/RTCP packets in s.rtpStreams[1].
	 *
	 * TODO: modify the application logic for dynamix stream management without
	 * assumption about how the RTP chunkiser works.
	 */
	var stream rtpStream
	logger.Info("StreamID/rtsStreams ", int(env.StreamID), len(s.rtpStreams))
	if len(s.rtpStreams) == 1 {
		stream = s.rtpStreams[0]
	} else if len(s.rtpStreams) == 2 {
		if env.StreamID == 2 || env.StreamID == 3 {
			logger.Info("Video")
			stream = s.rtpStreams[1]
		} else if env.StreamID == 0 || env.StreamID == 1 {
			logger.Info("Audio")
			stream = s.rtpStreams[0]
		} else {
			logger.WithFields(log.Fields{"stream": env.StreamID,
				}).Warn("Unknown stream ID")
			return
		}
	} else {
		logger.WithFields(log.Fields{"stream": env.StreamID,
			}).Warn("Unknown stream ID")
		//stream = s.rtpStreams[env.StreamID/2]
		return
	}
	rtp := env.StreamID%2 == 0
	if rtp {
		logger.Info("Dispatch RTP ", len(env.Content))
		stream.RTP <- env.Content
	} else {
		logger.Info("Dispatch RTCP ", len(env.Content))
		stream.RTCP <- env.Content
	}
}

func (s *pStream) CleanPartialAssemblies() {
	for seq, buf := range s.assemblyBuffers {
		if time.Now().Sub(buf.LastUpdated) > s.assemblyLifetime {
			delete(s.assemblyBuffers, seq)
		}
	}
}
