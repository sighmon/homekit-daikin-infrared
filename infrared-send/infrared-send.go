package main

import (
  "github.com/chbmuc/lirc"
  "log"
)

func main() {
  // Initialize with path to lirc socket
  ir, err := lirc.Init("/var/run/lirc/lircd")
  if err != nil {
    panic(err)
  }

  err = ir.Send("daikin POWER_ON")
  if err != nil {
    log.Println(err)
  }
}
