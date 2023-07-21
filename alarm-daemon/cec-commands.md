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
