package main

import (
	"encoding/json"
	"errors"
	"os"
)

type configRoot struct {
	HTTP    configHTTP     `json:"http"`
	Streams []configStream `json:"streams"`
}

type configHTTP struct {
	Listen string `json:"listen"`
}

//go:generate stringer -type configStreamKind config.go
type configStream struct {
	Description string                 `json:"description"`
	Listen      string                 `json:"listen"`
	Kind        configStreamKindString `json:"kind"`
}

type configStreamKind int

type configStreamKindString struct {
	Value configStreamKind
}

const (
	configStreamKindVideoWebM configStreamKind = iota
	configStreamKindAudioOpus
	configStreamKindAudioTest1
	configStreamKindVideoTest1
)

var errConfigInvalidStreamKind = errors.New("config: invalid stream kind")

func (s configStreamKindString) MarshalJSON() ([]byte, error) {
	m := map[configStreamKind][]byte{
		configStreamKindVideoWebM:  []byte(`"video-webm"`),
		configStreamKindAudioOpus:  []byte(`"audio-opus"`),
		configStreamKindAudioTest1: []byte(`"audio-test1"`),
		configStreamKindVideoTest1: []byte(`"video-test1"`),
	}
	v, ok := m[s.Value]
	if !ok {
		return nil, errConfigInvalidStreamKind
	}
	return v, nil
}

func (s *configStreamKindString) UnmarshalJSON(value []byte) error {
	m := map[string]configStreamKind{
		`"video-webm"`:  configStreamKindVideoWebM,
		`"audio-opus"`:  configStreamKindAudioOpus,
		`"audio-test1"`: configStreamKindAudioTest1,
		`"video-test1"`: configStreamKindVideoTest1,
	}
	v, ok := m[string(value)]
	if !ok {
		return errConfigInvalidStreamKind
	}
	s.Value = v
	return nil
}

var configDefault = configRoot{
	HTTP: configHTTP{
		Listen: ":8080",
	},
	Streams: []configStream{
		configStream{
			Description: "Channel 1",
			Listen:      ":60006",
			Kind:        configStreamKindString{Value: configStreamKindVideoWebM},
		},
		configStream{
			Description: "Radio 1",
			Listen:      ":60007",
			Kind:        configStreamKindString{Value: configStreamKindAudioOpus},
		}},
}

func configParseFile(path string) (*configRoot, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	config := configRoot{}
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
