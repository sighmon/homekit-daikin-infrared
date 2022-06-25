# HomeKit Daikin infrared accessory

An Apple HomeKit accessory for an infrared Daikin FTXS50KAVMA reverse cycle air conditioner remote control.

## TODO

- [ ] Decode IR codes for all of the functions we'd like to use
- [ ] Send those IR codes using the Go LIRC client
- [ ] Setup HAP Go library to send IR codes
- [ ] Setup GAP Go library to receive air conditioner commands

## Useful links

* Daikin IR protocol: https://github.com/blafois/Daikin-IR-Reverse
* IR transmitter: https://core-electronics.com.au/digital-ir-transmitter-module-arduino-compatible.html
* HAP Go library: https://github.com/brutella/hap
* Go client for LIRC (Linux Infrared Remote Control): https://github.com/chbmuc/lirc
