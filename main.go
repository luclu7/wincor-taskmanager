package main

import (
	"fmt"
	linuxproc "github.com/c9s/goprocinfo/linux"
	"go.bug.st/serial"
	"log"
	"strconv"
	"time"
)

const (
	ClearScreen = "\x1B\x5B\x32\x4A"

	MovePrefix = "\x1B\x5B"
	MoveSuffix = "H"

	SetCountryCodePrefix = "\x1B\x52"
)

type Connection struct {
	serial.Port
}

func (c *Connection) WriteData(data []byte) {
	n, err := c.Write(data)

	if err != nil {
		log.Panicln(err)
	}

	log.Println("Data written: ", n)
}

func (c *Connection) ClearScreen() {
	c.WriteData([]byte(ClearScreen))
}

func (c *Connection) MoveCursorTo(line, column int) {
	str := MovePrefix + fmt.Sprintf("%02d", line) + fmt.Sprintf("%02d", column) + MoveSuffix
	c.WriteData([]byte(str))
}

func (c *Connection) MoveCursorToHome() {
	c.MoveCursorTo(0, 0)
}

func main() {
	mode := &serial.Mode{
		BaudRate: 9600,
		Parity:   serial.OddParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open("/dev/ttyUSB0", mode)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	conn := &Connection{port}

	// clear screen
	conn.ClearScreen()
	conn.MoveCursorToHome()

	// loop forever
	for {
		conn.MoveCursorToHome()

		avg, err := linuxproc.ReadLoadAvg("/proc/loadavg")

		if err != nil {
			log.Fatal("loadavg read fail")
		}

		conn.WriteData([]byte("Load: " + strconv.FormatFloat(avg.Last1Min, 'f', 2, 64)))

		now := time.Now()

		conn.WriteData([]byte(" " + now.Format("15:04:05")))

		conn.MoveCursorTo(1, 0)

		// read memory info
		mem, err := linuxproc.ReadMemInfo("/proc/meminfo")
		if err != nil {
			log.Fatal("meminfo read fail")
		}

		usedRamGB := float64(mem.MemTotal-mem.MemFree) / 1024 / 1024
		totalRamGB := float64(mem.MemTotal) / 1024 / 1024

		conn.WriteData([]byte("Mem: " + strconv.FormatFloat(usedRamGB, 'f', 2, 64) + "/" + strconv.FormatFloat(totalRamGB, 'f', 2, 64) + " GB"))

		// wait for 300ms
		time.Sleep(300 * time.Millisecond)
	}
}
