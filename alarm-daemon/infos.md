# CEC Commands

## Check power status

When on:

```sh
$ echo "pow 0.0.0.0" | cec-client -s -d 1
opening a connection to the CEC adapter...
power status: on
```

When off:

```sh
$ echo "pow 0.0.0.0" | cec-client -s -d 1
opening a connection to the CEC adapter...
power status: standby
```

## Turn TV on

```sh
$ echo "on 0.0.0.0" | cec-client -s -d 1
opening a connection to the CEC adapter...
```

## Turn TV off

```sh
$ echo "standby 0.0.0.0" | cec-client -s -d 1
opening a connection to the CEC adapter...
```

## Switch to raspberry

```sh
$ echo "as 0.0.0.0" | cec-client -s -d 1
opening a connection to the CEC adapter...
```

## Systemd

```sh
$ cat /etc/systemd/system/alarm-daemon.service
[Unit]
Description=Divera Alarm Daemon
After=network.target
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=alarmdaemon
Group=alarmdaemon
WorkingDirectory=/home/alarmdaemon/.alarm-daemon/work/
EnvironmentFile=/home/alarmdaemon/.alarm-daemon/config/env
ExecStart=/usr/bin/alarm-daemon
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

## Config

```sh
$ which alarm-daemon
/usr/bin/alarm-daemon
```

```sh
$ cat .alarm-daemon/config/env
SUBSCRIPTION_NAME=divera-alarm
GOOGLE_APPLICATION_CREDENTIALS=/home/alarmdaemon/.alarm-daemon/config/service_account_key.json
SWITCH_ON_CMD=/home/alarmdaemon/.alarm-daemon/config/on.sh
SWITCH_OFF_CMD=/home/alarmdaemon/.alarm-daemon/config/off.sh
LINGER_TIME=20m
COMMAND_TIMEOUT=60s
LAST_ALARM_FILE=/home/alarmdaemon/.alarm-daemon/work/lastAlarm
```

User needs to be in group `video` to access the CEC device.
