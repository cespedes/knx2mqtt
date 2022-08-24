package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func (e Event) MarshalJSON() ([]byte, error) {
	var tmp struct {
		Time        time.Time
		Gateway     string
		Command     string
		Source      string
		Destination string
		Data        []byte
	}
	tmp.Time = e.Time.Truncate(time.Second)
	tmp.Gateway = e.Gateway
	tmp.Command = e.Command.String()
	tmp.Source = e.Source.String()
	tmp.Destination = e.Destination.String()
	tmp.Data = e.Data
	return json.Marshal(tmp)
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
			fmt.Printf("Network: Error parsing %v for %v (%s): %s\n", e.Data, e.Destination, nt.DPT, err.Error())
		} else {
			str += " " + nt.Names[0] + "=" + fmt.Sprint(dp)
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
		fmt.Printf("names: %v\n", config.Names)
	}

	client, err := NewMQTTClient(config.MQTTServer, MQTTPort)
	if err != nil {
		log.Fatal(err)
	}

	mqttChan1, err := client.Subscribe(fmt.Sprintf("%s/+/+/+", config.MQTTPrefix1))
	mqttChan2, err := client.Subscribe(fmt.Sprintf("%s/cmd", config.MQTTPrefix2))
	for {
		select {
		case msg := <-mqttChan1:
			var e Event
			_ = json.Unmarshal(msg.Payload, &e)

			// Log packet:
			s.Log(e)

			// Send prettified packet to MQTT:
			if e.Command != knx.GroupWrite && e.Command != knx.GroupResponse {
				break
			}
			if nt, ok := config.Addresses[e.Destination]; ok {
				dp, ok := dpt.Produce(nt.DPT)
				if !ok {
					fmt.Printf("Warning: unknown type %v in config file\n", nt.DPT)
					continue
				}
				if err := dp.Unpack(e.Data); err != nil {
					fmt.Printf("Network: Error parsing %v for %v (%s): %s\n", e.Data, e.Destination, nt.DPT, err.Error())
					continue
				}
				value := e.Time.Format("20060102-150405") + " " + fmt.Sprint(dp)
				for _, name := range nt.Names {
					topic := fmt.Sprintf("%s/%s", config.MQTTPrefix2, name)
					err = client.PublishRetain(topic, value)
					if err != nil {
						fmt.Printf("MQTT: Error publishing to %s: %v\n", topic, err.Error())
						continue
					}
				}
			}
		case msg := <-mqttChan2:
			cmd := strings.Split(string(msg.Payload), " ")
			var command knx.GroupCommand
			switch {
			case len(cmd) < 2,
				cmd[0] != "write" && cmd[0] != "read" && cmd[0] != "response",
				cmd[0] == "write" && len(cmd) != 3,
				cmd[0] == "read" && len(cmd) != 2,
				cmd[0] == "response" && len(cmd) != 3:
				fmt.Printf("received wrong MQTT command: %q\n", string(msg.Payload))
				continue
			case cmd[0] == "read":
				command = knx.GroupRead
			case cmd[0] == "response":
				command = knx.GroupResponse
			case cmd[0] == "write":
				command = knx.GroupWrite
			}
			groupName := cmd[1]
			var groupAddr cemi.GroupAddr
			var DPT string

			groupAddr, ok := config.Names[groupName]
			if !ok {
				fmt.Printf("received command to wrong group: %q\n", string(msg.Payload))
				continue
			}
			val := config.Addresses[groupAddr]
			DPT = val.DPT

			dp, ok := dpt.Produce(DPT)
			if !ok {
				fmt.Printf("Error: unknown type %v in config file\n", DPT)
				continue
			}
			if command == knx.GroupResponse || command == knx.GroupWrite {
				value := cmd[2]
				err := SetDPTFromString(dp, value)
				if err != nil {
					fmt.Printf("received wrong value in MQTT command: %q\n", string(msg.Payload))
					continue
				}
			}
			data := dp.Pack()
			if command == knx.GroupRead {
				data = []byte{0}
			}
			fmt.Printf("MQTT: received cmd %q: will send packet to %s\n", msg.Payload, groupAddr.String())
			topic := fmt.Sprintf("%s/cmd", config.MQTTPrefix1)
			groupEvent := knx.GroupEvent{Command: command, Destination: groupAddr, Data: data}
			event := Event{Time: time.Now(), GroupEvent: groupEvent}
			b, _ := event.MarshalJSON()
			client.Publish(topic, string(b))
		}
	}
}
