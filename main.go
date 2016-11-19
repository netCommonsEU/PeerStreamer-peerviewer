package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	log "github.com/Sirupsen/logrus"
)

var (
	logger      = log.New()
	config      *configRoot
	peerStreams []*pStream
	gPipelines  []*gPipeline
)

var (
	flagConfigPath     = kingpin.Flag("config", "Config file path.").Default("config.json").Short('c').String()
	flagDebug          = kingpin.Flag("debug", "Debug enabled.").Short('d').Bool()
	flagConfigTemplate = kingpin.Flag("template", "Print a template configuration file and exit.").Bool()
	//verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	//name    = kingpin.Arg("name", "Name of user.").Required().String()
)

func main() {
	kingpin.Parse()
	if *flagConfigTemplate {
		c, _ := json.MarshalIndent(configDefault, "", "    ")
		fmt.Println(string(c))
		return
	}
	if *flagDebug {
		logger.Level = log.DebugLevel
	}

	var err error
	config, err = configParseFile(*flagConfigPath)
	if err != nil {
		logger.Fatal("Unable to load config, ", err)
	}

	logger.Info("Peerviewer starting")

	err = initPeerStreams(config.Streams)
	if err != nil {
		logger.Fatal("Unable to start inbound streams")
	}

	initGPipelines(config.Streams)

	if err = http.ListenAndServe(config.HTTP.Listen, httpInit()); err != nil {
		logger.Panic("Error while listening on HTTP: ", err)
	}
}

func initPeerStreams(streamConfig []configStream) error {
	peerStreams = make([]*pStream, len(streamConfig))
	for i, cfg := range streamConfig {
		addr, err := net.ResolveUDPAddr("udp", cfg.Listen)
		if err != nil {
			logger.WithFields(log.Fields{
				"stream": i,
				"listen": cfg.Listen,
			}).Error("Unable to resolve listen address: ", err)
			return err
		}
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			logger.WithFields(log.Fields{
				"stream": i,
				"listen": cfg.Listen,
			}).Error("Unable to listen on UDP socket: ", err)
			return err
		}
		logger.WithFields(log.Fields{
			"port":   addr.Port,
			"stream": i,
			"kind":   cfg.Kind.Value.String(),
		}).Debug("Initialized inbound stream")
		var numMediaStreams int
		switch cfg.Kind.Value {
		case configStreamKindVideoWebM:
			//numMediaStreams = 2
			numMediaStreams = 2
		case configStreamKindVideoVP8:
			numMediaStreams = 1
		case configStreamKindAudioOpus:
			numMediaStreams = 1
		case configStreamKindAudioTest1:
			numMediaStreams = 1
		case configStreamKindVideoTest1:
			numMediaStreams = 1
		}
		s := newPStream(conn, numMediaStreams)
		peerStreams[i] = s
		go s.ListenInbound()
	}
	return nil
}

func initGPipelines(streamConfig []configStream) {
	gPipelines = make([]*gPipeline, len(streamConfig))
	for i, cfg := range streamConfig {
		var p gPipeline
		switch cfg.Kind.Value {
		case configStreamKindVideoWebM:
			p = newGPipelineVideoWebm(i)
			go p.ListenForData(peerStreams[i].RTPStreams())
		case configStreamKindVideoVP8:
			p = newGPipelineVideoVP8(i)
			go p.ListenForData(peerStreams[i].RTPStreams())
		case configStreamKindAudioOpus:
			p = newGPipelineAudioOpus(i)
			go p.ListenForData(peerStreams[i].RTPStreams())
		case configStreamKindAudioTest1:
			p = newGPipelineAudioTest1(i)
		case configStreamKindVideoTest1:
			p = newGPipelineVideoTest1(i)
		}
		gPipelines[i] = &p
	}
}
