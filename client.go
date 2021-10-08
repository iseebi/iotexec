package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

type MessageHandler func(Client, []byte)

type Client struct {
	Host           string
	Port           uint
	CaCert         x509.CertPool
	ClientID       string
	User           string
	Password       Password
	Topic          string
	ReconnectTime  uint
	MessageHandler MessageHandler
	mqttClient     *mqtt.Client
}

func (c *Client) waitRestart() {
	timer := time.NewTimer(time.Second * time.Duration(c.ReconnectTime))
	<-timer.C
	c.startClient()
}

func (c *Client) mqttMessageHandler(client mqtt.Client, message mqtt.Message) {
	defer message.Ack()
	if *c.mqttClient != client {
		return
	}
	log.Printf("receive topic:%s message: %v\n", message.Topic(), string(message.Payload()))
	c.MessageHandler(*c, message.Payload())
}

func (c *Client) mqttConnectionLostHandler(client mqtt.Client, err error) {
	if *c.mqttClient != client {
		return
	}
	log.Printf("Connection lost: %v\n", err)
	c.waitRestart()
}

func (c *Client) startClient() {
	password, err := c.Password.GetPassword()
	if err != nil {
		log.Fatalf("Password generation failed: %v", err)
	}
	address := fmt.Sprintf("tls://%s:%d", c.Host, c.Port)
	opts := mqtt.NewClientOptions().
		AddBroker(address).
		SetClientID(c.ClientID).
		SetProtocolVersion(4).
		SetKeepAlive(time.Second * 60).
		SetConnectionLostHandler(c.mqttConnectionLostHandler).
		SetUsername(c.User).
		SetPassword(password).
		SetTLSConfig(&tls.Config{
			RootCAs: &c.CaCert,
		})

	log.Printf("start connection to %s", address)
	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Printf("Connection failed: %v", token.Error())
		c.waitRestart()
	}

	subscribeToken := mqttClient.Subscribe(c.Topic, 0, c.mqttMessageHandler)
	if subscribeToken.Wait() && subscribeToken.Error() != nil {
		log.Printf("Subscribe failed: %v", subscribeToken.Error())
		c.waitRestart()
	}
	c.mqttClient = &mqttClient
	log.Printf("Connected")
}

func NewClient(host string, port uint, caCert x509.CertPool, clientID string, user string, password Password, topic string, messageHandler MessageHandler) *Client {
	return &Client{
		Host:           host,
		Port:           port,
		CaCert:         caCert,
		ClientID:       clientID,
		User:           user,
		Password:       password,
		Topic:          topic,
		ReconnectTime:  10,
		MessageHandler: messageHandler,
	}
}

func (c *Client) StartReceive() {
	c.startClient()
}

func (c *Client) StopReceive() {
	if *c.mqttClient == nil {
		return
	}
	mqttClient := *c.mqttClient
	c.mqttClient = nil
	mqttClient.Disconnect(0)
}
