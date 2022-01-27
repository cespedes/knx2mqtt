# knx2mqtt

knx2mqtt is a bidirectional KNX to MQTT bridge.

```sh
$ knx2mqtt -h
Usage of knx2mqtt:
  -knx value
        KNX Gateway (can be repeated)
  -mqtt string
        MQTT broker
  -mqtt-prefix string
        MQTT prefix to use (default "knx")
```

It connects to one or more KNX routers and to one MQTT broker.
It then listens to all the messages in the KNX network(s) and publishes
them as MQTT topics.  It also reads messages from MQTT and writes them
to KNX.

All the messages received from the KNX gateways are published to MQTT
with topic prefix/group-address and encoded as a JSON object like this:

	{"Time":"2022-01-25T16:46:00+01:00","Gateway":"192.168.1.50","Command":"Write","Source":"1.4.50","Destination":"5/0/27","Data":"AQ=="}

All the messages published by MQTT as topic prefix/command with the same
format are sent as KNX messages (ignoring Time and Source, and trying to
guess Gateway if not specified).
