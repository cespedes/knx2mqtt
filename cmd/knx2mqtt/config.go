package main

import (
	"flag"
	"fmt"
	"log"
)

type SliceOfStrings []string

func (s *SliceOfStrings) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprint([]string(*s))
}

func (s *SliceOfStrings) Set(value string) error {
	if s == nil {
		return fmt.Errorf("cannot set value of nil pointer")
	}
	*s = append(*s, value)
	return nil
}

type Config struct {
	Debug       bool
	KNXGateways SliceOfStrings
	MQTTServer  string
	MQTTPrefix  string
}

func ReadConfig() *Config {
	var config Config
	// var configFile string
	// flag.StringVar(&configFile, "config", "knx2mqtt.ini", "Config file to read")
	flag.BoolVar(&config.Debug, "debug", false, "Debugging")
	flag.Var(&config.KNXGateways, "knx", "KNX Gateway (can be repeated)")
	flag.StringVar(&config.MQTTServer, "mqtt", "", "MQTT server")
	flag.StringVar(&config.MQTTPrefix, "mqtt-prefix", "knx", "MQTT prefix to use")
	flag.Parse()

	if config.Debug {
		log.Printf("config = %+v\n", config)
	}

	if len(config.KNXGateways) == 0 {
		log.Fatalf("No KNX gateways specified")
	}
	if len(config.MQTTServer) == 0 {
		log.Fatalf("No MQTT server specified")
	}

	return &config
}
