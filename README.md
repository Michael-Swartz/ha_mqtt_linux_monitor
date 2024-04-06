### HA MQTT Linux Monitor


## Usage

```
$ ./ha_mqtt_monitor --help
Usage of ./ha_mqtt_monitor:
  -address string
        IP or FQDN of Mqtt broker (default "localhost")
  -port int
        Port Number (default 1883)
  -topic-prefix string
        Prefix for the mqtt topic. (default "computer-monitor")
```

## MQTT Topics

By default this will export the following metrics on the following topics:

| Metric | Topic |
| :------: | :------ |
| CPU Usage | /computer-monitor/cpu |
| Memory Usage | /computer-monitor/memory |
| CPU Temperature | /computer-monitor/temp/cpu |
| GPU Temperature | /computer-monitor/temp/gpu |