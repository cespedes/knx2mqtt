package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/cemi"
)

const (
	KNXDefaultPort = 3671
	KNXTimeout     = 3 * time.Minute // no messages in some time: probable error in connection
)

func toEvent(gw string, knxEvent knx.GroupEvent) Event {
	var event Event
	event.Time = time.Now().Truncate(time.Second)
	event.Gateway = gw
	event.Command = knxEvent.Command
	event.Source = knxEvent.Source
	event.Destination = knxEvent.Destination
	event.Data = knxEvent.Data
	return event
}

func (s *Server) KNX(gateways []string) (fromKNX chan Event, toKNX chan Event) {
	type clientAddresses struct {
		Client    knx.GroupTunnel
		Addresses []cemi.GroupAddr
	}
	gws := make(map[string]*clientAddresses)
	var mu sync.Mutex

	//conns := make(map[string]knx.GroupTunnel)

	for i, gw := range gateways {
		if !strings.Contains(gw, ":") {
			gateways[i] = fmt.Sprintf("%s:%d", gw, KNXDefaultPort)
		}
	}

	inChan := make(chan Event, 5)
	outChan := make(chan Event, 5)

	// Populate map before creating goroutines
	for _, gw := range gateways {
		gws[gw] = new(clientAddresses)
	}

	for _, gw := range gateways {
		go func(gwName string) {
			for {
				if s.Debug {
					log.Printf("Stablishing connection to KNX gateway %s...\n", gwName)
				}
				client, err := knx.NewGroupTunnel(gwName, knx.DefaultTunnelConfig)
				if err != nil {
					log.Printf("KNX: Could not connect to %q: %v", gw, err)
					if s.Debug {
						log.Printf("Sleeping %s...", KNXTimeout/4)
					}
					time.Sleep(KNXTimeout / 4)
					continue
				}
				defer client.Close()
				gws[gwName].Client = client

				knxChan := client.Inbound()

			knxReadLoop:
				for {
					knxEvent, ok := <-knxChan
					if !ok {
						log.Fatalf("error reading from KNX gateway %q", gwName)
					}
					if s.Debug {
						log.Printf("KNX: Received from %q: %v", gwName, knxEvent)
					}
					event := toEvent(gwName, knxEvent)
					outChan <- event
					for _, addr := range gws[gwName].Addresses {
						if addr == knxEvent.Destination {
							continue knxReadLoop
						}
					}
					// we only lock on writing because the other threads do not modify the slice
					mu.Lock()
					gws[gwName].Addresses = append(gws[gwName].Addresses, knxEvent.Destination)
					if s.Debug {
						log.Printf("gws[%q].Addresses has %d entries", gwName, len(gws[gwName].Addresses))
					}
					mu.Unlock()
				}
			}
		}(gw)
	}
	go func() {
		for {
			event := <-inChan
			addr := event.Destination
			var mask cemi.GroupAddr
			var client *knx.GroupTunnel
			mu.Lock()
		knxCheckLoop:
			for i := 0; i < 16; i++ {
				mask = ^cemi.GroupAddr(0)
				mask <<= i
				check := addr & mask
				for _, gw := range gateways {
					for _, gaddr := range gws[gw].Addresses {
						if check == (gaddr & mask) {
							log.Printf("(i=%d) wanna write to %s; %q sent packet to %s :)", i, addr, gw, gaddr)
							client = &gws[gw].Client
							break knxCheckLoop
						}
					}
				}
			}
			mu.Unlock()
			if client != nil {
				groupEvent := event.GroupEvent
				if s.Debug {
					log.Printf("sending %v", groupEvent)
				}
				err := client.Send(groupEvent)
				if err != nil {
					log.Fatalf("error writing to KNX gateway: %v", err)
				}

			}
		}
	}()

	return outChan, inChan
}
