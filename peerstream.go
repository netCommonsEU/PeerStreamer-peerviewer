package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/netCommonsEU/PeerStreamer-go-grapes"
	"github.com/netCommonsEU/PeerStreamer-go-ml"
)

type rtpStream struct {
	RTP  chan []byte
	RTCP chan []byte
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
		stream.rtpStreams[0].RTP = make(chan []byte)
		stream.rtpStreams[0].RTCP = make(chan []byte)
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
	if int(env.StreamID) >= len(s.rtpStreams)/2 {
		logger.WithFields(log.Fields{
			"stream": env.StreamID,
		}).Warn("Unknown stream ID")
		return
	}
	stream := s.rtpStreams[env.StreamID/2]
	rtp := env.StreamID%2 == 0
	if rtp {
		stream.RTP <- env.Content
	} else {
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
