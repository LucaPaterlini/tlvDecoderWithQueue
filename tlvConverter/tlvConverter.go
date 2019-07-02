// Package tlvconverter provide an encoder and decoder to tlv (type-length-value) format.
package tlvconverter

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"time"
)

// encodeVltFloat64 converts a float in a array of bytes containing the tlv encoding of the triplet.
func encodeTlvFloat64(numFloat float64) []byte {
	if numFloat == 0 {
		return []byte{FLOAT64, 1, 0}
	}
	numFloatB := make([]byte, 8)
	binary.BigEndian.PutUint64(numFloatB, math.Float64bits(numFloat))
	return append([]byte{FLOAT64, byte(8 - countUnusedBytes(numFloatB))}, numFloatB[countUnusedBytes(numFloatB):]...)
}

// encodeVltUint64 converts a uint64 in a array of bytes containing the tlv encoding of the triplet.
func encodeTlvUint64(num uint64) []byte {
	if num == 0 {
		return []byte{UINT64, 1, 0}
	}
	numB := make([]byte, 8)
	binary.BigEndian.PutUint64(numB, num)
	return append([]byte{UINT64, byte(8 - countUnusedBytes(numB))}, numB[countUnusedBytes(numB):]...)
}

// countUnusedBytes counts the bytes that can be freed to save space.
func countUnusedBytes(s []byte) (c int) {
	for _, b := range s {
		if b == 0 {
			c++
		}
	}
	return
}

// EncodeVltPackage encode a PacketType into its tlv array of bytes.
func EncodeVltPackage(p PacketType) (res []byte) {
	// added the package header
	res = []byte{PACKET}
	numT := make([]byte, 8)
	binary.BigEndian.PutUint64(numT, 69)
	res = append(res, numT[countUnusedBytes(numT):]...)
	// adding a
	res = append(res, encodeTlvUint64(p.A)...)
	// adding the float64 b,c,d,e
	for _, x := range []float64{p.B, p.C, p.D, p.E} {
		res = append(res, encodeTlvFloat64(x)...)
	}
	// add the boolean f
	var fB byte
	if p.F {
		fB++
	}
	res = append(res, []byte{BOOL, 1, fB}...)
	res[1] = byte(len(res) - 2)
	return
}

func bytesToUint64(b []byte) uint64 {
	size := int(b[0])
	array8bits := append(make([]byte, 8-size), b[1:1+size]...)
	return binary.BigEndian.Uint64(array8bits)
}

func bytesToFloat64(b []byte) float64 {
	return math.Float64frombits(bytesToUint64(b))
}

// DecodeTlvPackage decode a PacketType into its tlv array of bytes.
func DecodeTlvPackage(byteArray []byte) (p PacketType, err error) {
	// logic stay focused on a single item
	// ignore the len the go on each triplet
	if byteArray[0] != PACKET {
		err = errors.New("not a PacketType tlv bytes representation")
		return
	}
	// checking the len of the array
	if int(byteArray[1]) < 18 || int(byteArray[1])+2 != len(byteArray) {
		err = errors.New("package corrupted, too short")
		return
	}
	// converting byte array to uint64 a
	byteArray = byteArray[2:]
	p.A = bytesToUint64(byteArray[1:])

	// converting byte array to float64 b
	byteArray = byteArray[2+int(byteArray[1]):]
	p.B = bytesToFloat64(byteArray[1:])
	// converting byte array to float64 c
	byteArray = byteArray[2+int(byteArray[1]):]
	p.C = bytesToFloat64(byteArray[1:])
	// converting byte array to float64 d
	byteArray = byteArray[2+int(byteArray[1]):]
	p.D = bytesToFloat64(byteArray[1:])
	// converting byte array to float64 e
	byteArray = byteArray[2+int(byteArray[1]):]
	p.E = bytesToFloat64(byteArray[1:])

	// converting byte array to float64 f
	byteArray = byteArray[2+int(byteArray[1]):]
	if byteArray[2] > 0 {
		p.F = true
	}
	return
}

// DecodeTlvPackageThread wrapper for goroutine for the decodeTlvPackage
func DecodeTlvPackageThread(byteArray []byte, p *chan PacketType) {
	tmpPacket, err := DecodeTlvPackage(byteArray)
	if err == nil {
		*p <- tmpPacket
		close(*p)
		return
	}
	log.Print(err.Error())
	close(*p)
}

// FromFileToBytesChannel reads the the binary file at filename split it in junks
// then convert the junks of bytes in PacketTypes and send them over the channel byteArraysChannel
func FromFileToBytesChannel(filename string, bytesChunksToProcess *chan []byte) {
	// open the file
	f, err := os.Open(filename)
	if err != nil {
		close(*bytesChunksToProcess)
		log.Fatalf(err.Error())
		return
	}
	// prepare the scanner and the buffers
	scanner := bufio.NewScanner(f)
	defer f.Close()
	// any value greater than 2 to make it reach the first size byte
	scanner.Split(bufio.ScanBytes)
	bufferSingleItem := make([]byte, 0, 56)
	bytesToAdd := 2
	for scanner.Scan() {
		b := scanner.Bytes()

		// set the new size where to generate a new array o bytes
		if len(bufferSingleItem) == 1 {
			bytesToAdd = int(b[0]) + 1
		}

		bufferSingleItem = append(bufferSingleItem, b[0])
		bytesToAdd--
		if bytesToAdd == 0 {

			// decode the byte array into the package and add the result to the channel
			*bytesChunksToProcess <- bufferSingleItem
			bufferSingleItem = make([]byte, 0, 56)
		}
	}
	close(*bytesChunksToProcess)
	return
}

// FromBytesChanToMonitoredQueueMultiThread decode the junk of bytes into packet and add them to the Monitored Queue
// using multiple go routines but keeping the order serialized.
func FromBytesChanToMonitoredQueueMultiThread(bytesChunksToProcess *chan []byte, mq *MonitoredQueue) {
	// starting 1 go routine for each record and keeping the response chan to keep them in order
	arrayOfChan := make(chan chan PacketType, 2)

	// loading arrayOfchan with the chan of the returns DecodeTlvPackageThread
	go func() {
		for recordBytes := range *bytesChunksToProcess {
			responsePacket := make(chan PacketType, 1)
			go DecodeTlvPackageThread(recordBytes, &responsePacket)
			arrayOfChan <- responsePacket
		}
		close(arrayOfChan)
	}()
	// retrieving the results in a sequential order
	for decodeChan := range arrayOfChan {
		time.Sleep(mq.writeDelay)
		if pkt, ok := <-decodeChan; ok {
			mq.Add(pkt)
		}
	}
	mq.Close()
}

// PacketTypeToStringArray convert the content of the field of the PacketType into an array of strings.
func PacketTypeToStringArray(p PacketType) []string {
	v := reflect.ValueOf(p)
	values := make([]string, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		values[i] = fmt.Sprintf("%v", v.Field(i))
	}
	return values
}

// WriteToCSV writes the PacketTypes received over pChan into fileName csv.
func WriteToCSV(mq *MonitoredQueue, fileName string, errChan *chan error) {
	defer close(*errChan)
	f, err := os.Create(fileName)
	if err != nil {
		*errChan <- err
		close(*errChan)
		return
	}
	w := csv.NewWriter(f)
	for record, ok := mq.GetFirst(); ok; record, ok = mq.GetFirst() {
		time.Sleep(mq.readDelay)
		if err := w.Write(PacketTypeToStringArray(record)); err != nil {
			*errChan <- errors.New("error writing record to csv:" + err.Error())
			return
		}
	}
	// Write any buffered packet to the csv.
	w.Flush()
	if err := w.Error(); err != nil {
		*errChan <- err
		return
	}
	return
}
