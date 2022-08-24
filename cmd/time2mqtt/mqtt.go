package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/at-wat/mqtt-go"
)

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

func mqttPublish(client mqtt.Client, topic string, payload string) error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = client.Publish(ctx, &mqtt.Message{
		Topic:   topic,
		Payload: []byte(payload),
	})
	return err
}
