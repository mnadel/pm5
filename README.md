# PM5

A Bluetooth Low-Energy (BLE) Central device that connects to a Concept2 PM5 and receives workout summary data.

Ultimate goal is to upload workout data to Logbook (currently awaiting OAuth2 application creds from Concept2), i.e. a phone-less integration with Concept2's Logbook.

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
