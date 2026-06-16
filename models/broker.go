package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"

	"github.com/google/uuid"
	mqttServer "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/zuadi/webServer/utils"
)

type Broker struct {
	broker     *mqttServer.Server
	websocket  *listeners.Websocket
	proxy      *httputil.ReverseProxy
	brokerPort int
	wsPort     int
}

type MQTTMessage struct {
	ID            string `json:"uuid,omitempty"`
	RemoteAddress string `json:"remoteAddress,omitempty"`
	Topic         string `json:"topic"`
	Payload       string `json:"payload"`
}

func NewBroker(address string, brokerPort, wsPort int) (b *Broker, err error) {
	b = &Broker{}

	if address == "" {
		address = "127.0.0.1"
	}

	brokerPort, err = utils.GetFreePort(address, brokerPort)
	if err != nil {
		return
	}

	wsPort, err = utils.GetFreePort(address, wsPort)
	if err != nil {
		return
	}

	b.broker = mqttServer.New(&mqttServer.Options{
		InlineClient: true,
	})
	b.broker.AddHook(new(auth.AllowHook), nil)

	tcpListener := listeners.NewTCP(listeners.Config{
		ID:      "embedded-mqtt-tcp",
		Address: net.JoinHostPort(address, fmt.Sprint(brokerPort)),
	})

	err = b.broker.AddListener(tcpListener)
	if err != nil {
		return nil, err
	}

	b.websocket = listeners.NewWebsocket(listeners.Config{
		ID:      "embedded-mqtt-ws",
		Address: net.JoinHostPort(address, fmt.Sprint(wsPort)),
	})

	if err := b.broker.AddListener(b.websocket); err != nil {
		return nil, fmt.Errorf("Failed to add WebSocket listener: %v", err)
	}

	targetURL, err := url.Parse("http://" + b.websocket.Address())

	if err != nil {
		return nil, fmt.Errorf("Failed to parse internal broker URL: %v", err)
	}

	b.proxy = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			// SetURL handles modifying scheme and host safely
			pr.SetURL(targetURL)

			// Safely rewrite incoming path "/mqtt" to root "/" for the internal broker
			pr.Out.URL.Path = "/"

			// Explicitly preserve the inbound Host header for WebSockets
			pr.Out.Host = pr.In.Host

			// Securely manages X-Forwarded-For to prevent client IP spoofing
			pr.SetXForwarded()
		},
	}

	go func() {
		if err = b.broker.Serve(); err != nil {
			log.Fatalf("mqtt error: %v", err)
		}
	}()

	return
}

func (b *Broker) ServeWS(w http.ResponseWriter, r *http.Request) {
	b.proxy.ServeHTTP(w, r)
}

func (b *Broker) Publish(topic string, payload any) (err error) {
	if topic == "" {
		return errors.New("missing topic")
	} else if payload == "" {
		return errors.New("missing payload")
	}

	var p []byte

	switch v := payload.(type) {
	case []byte:
		p = v
	case string:
		p = []byte(v)
	case MQTTMessage:
		p, err = json.Marshal(v)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("type '%v' not supported", reflect.TypeOf(v))
	}

	return b.broker.Publish(topic, p, false, 0)
}

func (b *Broker) Subscribe(topic string, cb func(msg MQTTMessage)) error {
	if topic == "" {
		return errors.New("missing topic")
	}
	id, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	err = b.broker.Subscribe(topic, int(id.ID()), func(cl *mqttServer.Client, sub packets.Subscription, pk packets.Packet) {
		if cb == nil {
			return
		}

		msg := MQTTMessage{
			ID:            cl.ID,
			RemoteAddress: pk.Origin,
			Topic:         pk.TopicName,
			Payload:       string(pk.Payload),
		}
		cb(msg)
	})
	return err
}
