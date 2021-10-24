# PM5

A Bluetooth Low-Energy (BLE) Central device that connects to a Concept2 PM5 and receives workout summary data and transmits it to [C2's Logbook](https://log.concept2.com/).

# Installing

1. Clone this repo
1. Run `go install`
1. Update `pm5.service` to point to the `pm5` binary and set your Concept2 client secret
1. Run `sudo make setup`

NB you might also need to pair your device with your PM5. On my Pi it looked something like this:

```
> sudo bluetoothctl
[bluetooth]# scan on
** observe your PM5 fly by **
[bluetooth]# pair <addr>
[bluetooth]# trust <addr>
```

# Configuration

## `ble_watchdog_workout_deadline`

It's possible to connect to the PM5 but never receive Workout data nor a disconnect event. In this case, we'd indefinitely await Workout data.

This setting should be longer than your expected workout. We'll start a watchdog timer when we connect to the PM5 and if we don't receive a Workout summary before this deadline, we'll reset ourselves.

My usage of the Concept2 is for a 5-10 min warmup and cooldown, so I use a relatively low value for this (15m).

# Auth

OAuth callbacks are handled by a [Cloudflare Worker](https://workers.cloudflare.com/) deployed to https://auth.pm5-book.workers.dev/c2.

Its source is available at https://github.com/mnadel/pm5-auth.

First, generate a link to authenticate PM5 Book:

```
> pm5 --authurl
```

After navigating to the link shown and authorizing this application, you'll be shown a command to run. It'll look something like this:

```
> pm5 --auth xxxyyy:abc123
```

And with that, PM5 Book will have everything it needs to update Logbook on your behalf!

Note that the refresh token is valid for a year. With each sync we'll try obtaining a new refresh token, so you shouldn't have to re-authenticate unless you haven't used PM5-Book in over a year. If your auth tokens do expire, you cat visit `pm5 --authurl` to get new tokens and re-invoke `pm5 --auth`.
