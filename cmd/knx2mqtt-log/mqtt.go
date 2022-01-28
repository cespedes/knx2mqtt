package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/at-wat/mqtt-go"
)

type MQTTClient struct {
	mutex         sync.Mutex
	client        mqtt.ClientCloser
	mux           *mqtt.ServeMux
	subscriptions []struct {
		topic string
		ch    chan *mqtt.Message
	}
}

func mqttConnect(server string, port int) (mqtt.ClientCloser, error) {
	addr := fmt.Sprintf("mqtt://%s:%d", server, port)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mqtt.DialContext(ctx, addr)
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	_, err = client.Connect(ctx, fmt.Sprint(rand.Uint64()))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewMQTTClient(server string, port int) (*MQTTClient, error) {
	var m MQTTClient
	var err error

	m.client, err = mqttConnect(server, port)
	if err != nil {
		return nil, err
	}
	m.mux = &mqtt.ServeMux{}
	m.client.Handle(m.mux)

	go func() {
		for {
			select {
			case <-m.client.Done():
				// Connection closed; will have to reconnect
				var client mqtt.ClientCloser
				var err error
				for {
					time.Sleep(500 * time.Millisecond)
					log.Println("MQTT: connection closed to server.  Reconnecting...")
					client, err = mqttConnect(server, port)
					if err == nil {
						break
					}
					log.Printf("MQTT: error connecting to %s: %v", server, err)
					time.Sleep(500 * time.Millisecond)
				}
				mux := &mqtt.ServeMux{}
				client.Handle(mux)
				m.mutex.Lock()
				m.client = client
				m.mux = mux
				for _, sub := range m.subscriptions {
					mux.HandleFunc(sub.topic, func(m *mqtt.Message) {
						sub.ch <- m
					})
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					defer cancel()
					client.Subscribe(ctx, mqtt.Subscription{Topic: sub.topic})
				}
				m.mutex.Unlock()
			}
		}
	}()
	return &m, nil
}

func (m *MQTTClient) Publish(topic string, payload string) error {
	var err error

	m.mutex.Lock()
	client := m.client
	m.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = client.Publish(ctx, &mqtt.Message{
		Topic:   topic,
		Payload: []byte(payload),
	})
	return err
}

func (m *MQTTClient) Subscribe(topic string) (chan *mqtt.Message, error) {
	ch := make(chan *mqtt.Message)

	m.mutex.Lock()
	mux := m.mux
	client := m.client
	m.subscriptions = append(m.subscriptions, struct {
		topic string
		ch    chan *mqtt.Message
	}{topic, ch})
	m.mutex.Unlock()

	mux.HandleFunc(topic, func(m *mqtt.Message) {
		ch <- m
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.Subscribe(ctx, mqtt.Subscription{Topic: topic})
	if err != nil {
		return nil, err
	}

	return ch, nil
}
