package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

const (
	MQTTPort = 1883
)

type Event struct {
	Time        time.Time
	Gateway     string
	Command     string
	Source      string
	Destination string
	Data        []byte
}

/*
func FromJSON(b []byte) knx.GroupEvent {
	var my MyGroupEvent
	var ret knx.GroupEvent

	json.Unmarshal(b, &my)

	switch my.Command {
	case "read", "Read":
		ret.Command = knx.GroupRead
	case "write", "Write":
		ret.Command = knx.GroupWrite
	case "response", "Response":
		ret.Command = knx.GroupResponse
	}
	// ret.Source, _ = cemi.NewIndividualAddrString(my.Source)
	ret.Destination, _ = cemi.NewGroupAddrString(my.Destination)
	ret.Data = my.Data
	return ret
}
*/

func (s *Server) MQTT(server string, prefix string) (fromMQTT chan Event, toMQTT chan Event) {
	in := make(chan Event, 5)
	out := make(chan Event, 5)

	go func() {
		if s.Debug {
			log.Printf("MQTT: Connecting to server %q...", server)
		}
		client, err := NewMQTTClient(server, MQTTPort)
		if err != nil {
			log.Fatalf("MQTT: Could not connect to %q: %v", server, err)
		}
		if s.Debug {
			log.Printf("MQTT: Connected.")
		}

		subTopic := fmt.Sprintf("%s/#", prefix)
		mqttChan, err := client.Subscribe(fmt.Sprintf("%s/cmd", prefix))
		if err != nil {
			log.Fatalf("MQTT: subscribing to %s: %v", subTopic, err)
		}

		for {
			select {
			case m := <-mqttChan:
				log.Printf("MQTT: got MQTT packet: %v", m)
				out <- Event{}
				log.Printf("MQTT: packet sent to program")
			case event := <-in:
				topic := fmt.Sprintf("%s/%v", prefix, event.Destination)
				b, _ := json.Marshal(event)
				err = client.Publish(topic, string(b))
				if err != nil {
					log.Printf("MQTT: publishing to %s: %s", topic, err.Error())
				}
			}
		}
	}()
	return out, in
}

type Server struct {
	Debug bool
}

func main() {
	config := ReadConfig()

	s := &Server{}
	s.Debug = config.Debug

	// get channels to read and write to KNX network
	if s.Debug {
		log.Printf("connecting to KNX gateways %v\n", config.KNXGateways)
	}
	fromKNX, toKNX := s.KNX(config.KNXGateways)

	// get channels to read and write MQTT messages
	if s.Debug {
		log.Printf("connecting to MQTT server %s\n", config.MQTTServer)
	}
	fromMQTT, toMQTT := s.MQTT(config.MQTTServer, config.MQTTPrefix)

	if s.Debug {
		log.Println("waiting for packets...")
	}
	for {
		select {
		case m := <-fromKNX:
			if s.Debug {
				log.Printf("KNX -> MQTT: %v", m)
			}
			toMQTT <- m
		case m := <-fromMQTT:
			log.Printf("MQTT -> KNX: %v", m)
			toKNX <- m
		}
	}
}
