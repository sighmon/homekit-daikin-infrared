# HomeKit Daikin infrared accessory

An Apple HomeKit accessory for an infrared Daikin FTXS50KAVMA reverse cycle air conditioner remote control.

## Hardware

* Duinotech infrared receiver XC-4427 ([Jaycar](https://www.jaycar.com.au/arduino-compatible-infrared-receiver-module/p/XC4427))
* Duinotech infrared transmitter XC-4426 ([Jaycar](https://www.jaycar.com.au/arduino-compatible-infrared-transmitter-module/p/XC4426))

### Wiring

| Raspberry Pi pin | Infrared receiver pin |
| - | - |
| `9` Ground | `-` |
| `11` GPIO17 | `s` |
| `17` 3.3V | `middle` |

## Software

* Install `LIRC`: `sudo apt install lirc`

To detect IR codes, run: `mode2`

If pressing your remote doesn't output anything, try: `mode2 --driver default`

You can also set this as a default by editing `/etc/lirc/lirc_options.conf`

```conf
# /etc/lirc/lirc_options.conf
driver = default
device = /dev/lirc0
```

## Decoding Daikin IR codes

Find my collection of Daikin FTXS50KAVMA infrared codes in the [/codes](/codes) folder.

This was the process to create these text files:

* Run: `mode2 --driver default > power_on.txt`
* Press the power on button on your remote once
* `<control> + c` to quit

I modified [Ben's code](https://www.time0ut.org/blog/posts/aircooling_automation/) a little to convert those LIRC pulse widths into binary strings.

Find the decoder at [/codes/decode.py](/codes/decode.py):

* Run: `python3 decode.py power_on.txt`

Sample output: `00000BA1000100001011011111001000000000010100011000000000000000011101011BA1000100001011011111001000000000001000010000010011101110011111000BA10001000010110111110010000000000000000001001001001001100000000000000010100000000000000000110000000000110000000000000000010000011000000000000000000101010E`

Where:

* `A` = start of line
* `B` = end of line
* `E` = end of command

### Daikin FTXS50KAVMA IR codes

Here are the hex codes from the binary strings `decode.py` output:

**Power on**

```hex
00 00
88 5B E4 00 A3 00 00 EB
88 5B E4 00 42 09 DC F8
11 0B 7C 80 00 12 49 80 00 A0 00 0C 00 C0 00 10 60 00 05
```

**Power off**

```hex
00 00
88 5B E4 00 A3 00 00 EB
88 5B E4 00 42 A9 DC 24
11 0B 7C 80 00 02 49 80 00 A0 00 0C 00 C0 00 10 60 00 19
```

**Temperature up**

```hex
00 00
88 5B E4 00 A3 00 00 EB
88 5B E4 00 42 E9 DC 64
11 0B 7C 80 00 02 45 80 00 A0 00 0C 00 C0 00 10 60 00 15
```

**Temperature down**

```hex
00 00
88 5B E4 00 A3 00 00 EB
88 5B E4 00 42 19 DC E4
11 0B 7C 80 00 02 49 80 00 A0 00 0C 00 C0 00 10 60 00 19
```

The lengths seem to match these [reversed Daikin codes for a `ARC470A1` remote](https://github.com/blafois/Daikin-IR-Reverse#protocol-documentation).

## TODO

- [x] Decode IR codes for all of the functions we'd like to use
- [ ] Send those IR codes using the Go LIRC client
- [ ] Setup HAP Go library to send IR codes
- [ ] Setup GAP Go library to receive air conditioner commands

## Useful links

* Daikin IR protocol: https://github.com/blafois/Daikin-IR-Reverse
* IR transmitter: https://core-electronics.com.au/digital-ir-transmitter-module-arduino-compatible.html
* HAP Go library: https://github.com/brutella/hap
* Go client for LIRC (Linux Infrared Remote Control): https://github.com/chbmuc/lirc
