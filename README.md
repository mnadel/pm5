# PM5

A Bluetooth Low-Energy (BLE) Central device that connects to a Concept2 PM5 and receives workout summary data.

Ultimate goal is to upload workout data to Logbook (currently awaiting OAuth2 application creds from Concept2) -- i.e. phone-less integration with Concept2's Logbook.

# Installing

1. Clone this repo
1. Run `go install`
1. Update `pm5.service` to point to the `pm5` binary
1. Update `pm5.yml` and specify your PM5's address in `pm5_device_addr`
1. Run `sudo make install`
