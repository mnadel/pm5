# PM5

A Bluetooth Low-Energy (BLE) Central device that connects to a Concept2 PM5 and receives workout summary data and transmits it to [C2's Logbook](https://log.concept2.com/).

# Installing

1. Clone this repo
1. Run `go install`
1. Update `pm5.service` to point to the `pm5` binary
1. Run `sudo make setup`

NB you'll also need to pair your device with your PM5. On my Pi it looked something like this:

```
> sudo bluetoothctl
[bluetooth]# scan on
** observe your PM5 fly by **
[bluetooth]# pair <addr>
[bluetooth]# trust <addr>
```

git@github.com:mnadel/pm5-auth.git# Auth

OAuth callbacks are handled by https://github.com/mnadel/pm5-auth
