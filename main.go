package main

import (
	"crypto/rsa"
	"flag"
	jwt "github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

func pathParse(text string) map[string]string {
	items := strings.Split(text, "/")
	result := map[string]string{}
	for i := 0; i < len(items); i += 2 {
		if i+1 >= len(items) {
			break
		}
		key := items[i]
		value := items[i+1]
		result[key] = value
	}
	return result
}

func readPrivateKey(privateKeyFile string) (*rsa.PrivateKey, error) {
	signBytes, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(signBytes)
}

func main() {
	host := flag.String("Host", "mqtt.googleapis.com", "Cloud IoT MQTT Host")
	port := flag.Uint("port", 8883, "Cloud IoT MQTT Port")
	caCert := flag.String("caCert", "https://pki.goog/roots.pem", "MQTT CA Certificate File (if specified url, fetch)")
	mqttClientId := flag.String("mqttClientId", "", "MQTT Client ID")
	mqttUser := flag.String("mqttUser", "mqtt", "MQTT username")
	mqttPassword := flag.String("mqttPassword", "", "MQTT username (if set, it takes precedence over JWT setting)")
	mqttTopic := flag.String("mqttTopic", "", "MQTT topic")
	jwtPrivateKey := flag.String("jwtPrivateKey", "", "JWT private key file")
	jwtAudience := flag.String("jwtAudience", "", "JWT Audience (GCP Project ID in Cloud IoT Core, if not set, try to parse from Client ID)")
	command := flag.String("command", "", "execute command")
	flag.Parse()

	if *mqttTopic == "" {
		log.Fatalf("Must provide MQTT topic by --mqttTopic option")
	}
	if *mqttClientId == "" {
		log.Fatalf("Must provide MQTT client ID by --mqttClientId option")
	}
	if *caCert == "" {
		log.Fatalf("Must provide MQTT certificate by --caCert option")
	}

	// create jwt or use password
	var password Password
	if *mqttPassword == "" {
		if *jwtPrivateKey == "" {
			log.Fatalf("Must provide JWT private key by --jwtPrivateKey option")
		}

		audience := *jwtAudience
		if *jwtAudience == "" {
			pathItems := pathParse(*mqttClientId)
			if pathItems["projects"] == "" {
				log.Fatalf("Must provide JWT audience by --jwtAudience option")
			}
			audience = pathItems["projects"]
		}
		privateKey, err := readPrivateKey(*jwtPrivateKey)
		if err != nil {
			log.Fatalf("Read Private Key failed: %v", err)
		}
		password = NewJwtPassword(privateKey, audience)
	} else {
		password = NewRawPassword(*mqttPassword)
	}

	certPool, err := GetCertPool(*caCert)
	if err != nil {
		log.Fatalf("Read Certificate failed: %v", err)
	}

	messageCh := make(chan []byte)
	messageHandler := func(client Client, payload []byte) {
		messageCh <- payload
	}

	client := NewClient(
		*host, *port, *certPool,
		*mqttClientId, *mqttUser, password,
		*mqttTopic, messageHandler)
	client.StartReceive()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	for {
		select {
		case payload := <-messageCh:
			log.Printf("received action payload: %v\n", string(payload))
			if err := runCommand(*command, payload); err != nil {
				log.Printf("Execute error: %v", err)
			}
		case <-signalCh:
			log.Printf("Interrupt detected.\n")
			client.StopReceive()
			return
		}
	}
}

func runCommand(command string, payload []byte) error {
	cmd := exec.Command(command)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if _, err = stdin.Write(payload); err != nil {
		return err
	}
	if err = stdin.Close(); err != nil {
		return err
	}
	_, err = cmd.Output()
	return err
}
