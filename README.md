# iotexec

Execute specific command when receive MQTT message from Google Cloud IoT Core

## How to use

[Register a device in Cloud IoT Core](https://cloud.google.com/iot/docs/how-tos/devices) and prepate thease parameters

- RSA Private Key PEM File
- MQTT Client ID: `projects/${GCP_PROJECT}/locations/${GCP_REGION}/registries/${IOT_REGISTRY}/devices/${DEVICE_ID}`
- MQTT Topic
    - [Device State](https://cloud.google.com/iot/docs/how-tos/mqtt-bridge#setting_device_state): `/devices/${DEVICE_ID}/state`
    - [Command](https://cloud.google.com/iot/docs/how-tos/commands#receiving_a_command): `/devices/${DEVICE_ID}/commands/#` (it's required to wildcard)

then start iotexec

```
$ iotexec \
  --mqttClientId "projects/${GCP_PROJECT}/locations/${GCP_REGION}/registries/${IOT_REGISTRY}/devices/${DEVICE_ID}" \
  --mqttTopic "/devices/${DEVICE_ID}/commands/#" \
  --jwtPrivateKey rsa_private.pem \
  --command target_command.sh
```

## License

MIT