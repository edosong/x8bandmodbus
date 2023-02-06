package main

// The {band OA} x 8 ch.
// Data: 32bit float,
// Data (each byte) order: LittleEndian -> need to change to BigEdian
// Date:20230206-

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/goburrow/modbus"
)

const (
	RegQuantity     uint16 = 96 // the driver supports up to 125, need divide into two parts 96,96->192
	DefaultHost            = "192.168.179.5"
	DefaultPort            = "502" //standard
	DefaultRate            = 60    //second
	MinRate                = 10
	DefaultDataNums        = 10 //
	MaxDataNums            = 1000
)

var startBandOAAddress = []uint16{0xAA51, 0xAAB1} // 43601, ~+96=43697

func main() {

	// get device host (url or ip address) and port from the command line
	var (
		host     string
		port     string
		rate     int64
		dataNums int64
	)

	flag.StringVar(&host, "host", DefaultHost, "Slave device host (url or ip address)")
	flag.StringVar(&port, "port", DefaultPort, fmt.Sprintf("Slave device port (default:%s)", DefaultPort))
	flag.Int64Var(&rate, "rate", DefaultRate, "Data collection rate in Second. > 10 required.")
	flag.Int64Var(&rate, "nums", DefaultDataNums, fmt.Sprintf("Number (Default:%d Max:%d) of data to collect.", DefaultDataNums, MaxDataNums))

	flag.Parse()
	if rate < MinRate {
		rate = MinRate
	}
	if dataNums > MaxDataNums {
		dataNums = MaxDataNums
	}
	mbHandler := modbus.NewTCPClientHandler(host + ":" + port)
	mbHandler.Timeout = 10 * time.Second
	mbHandler.SlaveId = 1

	var err error

	if err = mbHandler.Connect(); err != nil {
		log.Fatal("Connect error:", err)
	}
	defer mbHandler.Close()

	client := modbus.NewClient(mbHandler)
	printOACsvHeader()

	//-- processing 8 chan per specified rate
	// x8Data := make([]byte, 64)
	for i := 0; i < 100; i++ {
		readX8BandOA(client)
		fmt.Println()
		time.Sleep(time.Duration(rate) * time.Second)
	}
}

func readX8BandOA(client modbus.Client) {
	fmt.Printf("%s,", time.Now().Format("2006-01-02 15:04:05"))
	for q := 0; q < 2; q++ { //2 parts of 192 data
		x8Data, err := client.ReadHoldingRegisters(startBandOAAddress[q], RegQuantity)

		if err != nil {
			fmt.Println("Read holding reg error.", err)
		}
		if len(x8Data) != int(RegQuantity*2) {
			fmt.Println("x8Data length error.", len(x8Data))
		}

		for i := 0; i < 4; i++ { //4 chan x 2 time for 8ch
			for j := 0; j < 7; j++ { //speed + 6 bands + 5 unused -> 12ch = 48bytes
				dout := uint32(x8Data[0+j*4+i*48]) | uint32(x8Data[1+j*4+i*48])<<8 | uint32(x8Data[2+j*4+i*48])<<16 | uint32(x8Data[3+j*4+i*48])<<24
				fout := math.Float32frombits(dout)
				fmt.Printf("%.2f, ", fout)
			}
		}
	}
}

func printOACsvHeader() {
	fmt.Printf("X8II Band OA Modbus Data\n----------------------\n")
	fmt.Print("Time, ")
	for i := 1; i <= 8; i++ {
		fmt.Printf("Speed,A(ch%d), B(ch%d),C(ch%d),D(ch%d),E(ch%d),F(ch%d),", i, i, i, i, i, i)
	}
	fmt.Println()
}
