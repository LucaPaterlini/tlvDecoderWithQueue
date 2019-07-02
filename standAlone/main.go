// Package main contains the implementation of a standalone monitored message queue.
package main

import (
	"../tlvConverter"
	"flag"
	"log"
	"os"
	"time"
)

var sizeMonitoredQueue = flag.Int("bufsize", 3, "size of the channel of the monitored queue")
var timeMonitoredQueue = flag.Int("secload", 1, "frequency of the load checks on the queue")
var delayReader = flag.Int("dr", 0, "frequency of the load checks on the queue")

// the input file path is provided as arg while bufnum is configured
func main() {
	flag.Parse()
	// check the usage
	if len(flag.Args()) < 2 {
		log.Printf("Usage: %s [-dr=0 -secload=1 -bufsize=3] \"pathInputFile\" \"pathOutputCsv\"",os.Args[0])
		os.Exit(0)
	}

	InputFilePath := flag.Args()[0]
	pathOutputCsv := flag.Args()[1]
	// dividing the input read in chunks and sending them over a byte chan
	tmpBytesChunksToProcess := make(chan []byte, 2)
	go tlvconverter.FromFileToBytesChannel(InputFilePath, &tmpBytesChunksToProcess)
	// converting the bytes chunks into packets and adding them to the monitoredQueue
	nMq := tlvconverter.NewMonitoredQueue(*sizeMonitoredQueue)
	go tlvconverter.FromBytesChanToMonitoredQueueMultiThread(&tmpBytesChunksToProcess, &nMq)

	// start the monitor that checks if the queue is full increase the delay if its empty
	// decrease the delay
	ticker := time.NewTicker(time.Duration(*timeMonitoredQueue)* time.Second)
	go func() {
		for {
			<-ticker.C
			nMq.MonitorLoad()
		}
	}()
	// setting the delay for the reader
	nMq.SetReadDelay(time.Duration(*delayReader) * time.Second)
	// write on the log all the errors during the writing to the csv
	errChan := make(chan error, 1)

	tlvconverter.WriteToCSV(&nMq, pathOutputCsv, &errChan)
	for err := range errChan {
		log.Fatalf(err.Error())
	}
}
