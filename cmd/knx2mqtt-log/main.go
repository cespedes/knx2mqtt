package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
	"github.com/vapourismo/knx-go/knx/dpt"
)

const (
	MQTTPort = 1883
)

var config *Config

type Server struct {
	Debug bool

	logFile     *os.File
	logFileName string
}

type Event struct {
	Time    time.Time
	Gateway string
	knx.GroupEvent
}

func (e *Event) UnmarshalJSON(b []byte) error {
	var tmp struct {
		Time        time.Time
		Gateway     string
		Command     string
		Source      string
		Destination string
		Data        []byte
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	e.Time = tmp.Time
	e.Gateway = tmp.Gateway
	switch tmp.Command {
	case "read", "Read":
		e.Command = knx.GroupRead
	case "write", "Write":
		e.Command = knx.GroupWrite
	case "response", "Response":
		e.Command = knx.GroupResponse
	}
	e.Source, err = cemi.NewIndividualAddrString(tmp.Source)
	if err != nil {
		return err
	}
	e.Destination, err = cemi.NewGroupAddrString(tmp.Destination)
	if err != nil {
		return err
	}
	e.Data = tmp.Data
	return nil
}

func (e Event) String() string {
	str := fmt.Sprintf("%s <%s> %s: %s %s=%v",
		e.Time.Format("2006-01-02 15:04:05"),
		e.Gateway,
		e.Command,
		e.Source,
		e.Destination,
		e.Data,
	)
	if devStr, ok := config.Devices[e.Source]; ok {
		str += " " + devStr
	}
	if nt, ok := config.Addresses[e.Destination]; ok {
		dp, ok := dpt.Produce(nt.DPT)
		if !ok {
			fmt.Printf("Warning: unknown type %v in config file\n", nt.DPT)
			dp = new(UnknownDPT)
		}
		if err := dp.Unpack(e.Data); err != nil {
			fmt.Printf("Network: Error parsing %v for %v\n", e.Data, e.Destination)
		} else {
			str += " " + nt.Name + "=" + fmt.Sprint(dp)
		}
	}

	//group, _ := cemi.NewGroupAddrString(e.Destination)
	return str
}

func main() {
	var s Server
	flag.BoolVar(&s.Debug, "debug", false, "debugging info")
	configFile := flag.String("config", "knx.cfg", "config file")
	flag.Parse()

	var err error
	config, err = ReadConfig(*configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if config.MQTTServer == "" {
		log.Fatal("No MQTT server specified")
	}
	if s.Debug {
		fmt.Printf("devices: %v\n", config.Devices)
		fmt.Printf("addresses: %v\n", config.Addresses)
	}

	client, err := NewMQTTClient(config.MQTTServer, MQTTPort)
	if err != nil {
		log.Fatal(err)
	}

	mqttChan, err := client.Subscribe(fmt.Sprintf("%s/#", config.MQTTPrefix1))
	for {
		msg := <-mqttChan
		var e Event
		_ = json.Unmarshal(msg.Payload, &e)
		s.Log(e)
	}
}
