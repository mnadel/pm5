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

# Auth

OAuth callbacks are handled by a [Cloudflare Worker](https://workers.cloudflare.com/) deployed to https://auth.pm5-book.workers.dev/c2.

Its source is available at https://github.com/mnadel/pm5-auth.

First, generate a link to authenticate PM5 Book:

```
> ./pm5 -auth
```

After navigating to that link and authorizing this application, you'll be shown a command to run. It'll look something like this:

```
> ./pm5 -access xxxyyy -refresh abc123
```

And with that, PM5 Book will have everything it needs to update Logbook on your behalf!

Note that the refresh token is valid for a year, so you'll eventually need to run through the above auth flow again.
