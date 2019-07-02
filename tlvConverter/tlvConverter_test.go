package tlvconverter

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestEncodeTlvPackage(t *testing.T) {
	for _, test := range testCasesTlvconverter {
		if byteArray := EncodeVltPackage(test.input); !reflect.DeepEqual(byteArray, test.expected) {
			t.Fatalf("FAIL: %s: EncodeVltPackage(%v) =\n\t   %v\n, want %v.", test.description, test.input, byteArray, test.expected)
		}
		log.Printf("PASS: EncodeVltPackage :%s\n", test.description)
	}
}

func TestDecodeTlvPackageSuccess(t *testing.T) {
	for _, test := range testCasesTlvconverter {
		if pk, _ := DecodeTlvPackage(test.expected); !reflect.DeepEqual(pk, test.input) {
			t.Fatalf("FAIL: %s: DecodeTlvPackage(%v) =\n\t   %v\n, want %v.", test.description, test.expected, pk, test.input)
		}
		log.Printf("PASS: DecodeTlvPackage : %s\n", test.description)
	}
}

func TestDecodeTlvPackageFailure(t *testing.T) {
	for _, test := range testCasesTlvconverterFailure {
		if _, err := DecodeTlvPackage(test.input); !reflect.DeepEqual(err, test.expectedError) {
			t.Fatalf("FAIL: %s: DecodeVltPackage(%v) =\n\t  expected Error  %v\n, want %v.",
				test.description, test.input, err, test.expectedError)
		}
		log.Printf("PASS: DecodeVltPackage Failure: %s\n", test.description)
	}
}

func TestDecodeTlvPackageThread(t *testing.T) {
	for _, test := range testCasesTlvconverter {
		tmpChan := make(chan PacketType, 1)
		go DecodeTlvPackageThread(test.expected, &tmpChan)
		tmpPkt, ok := <-tmpChan
		if !ok || !reflect.DeepEqual(tmpPkt, test.input) {
			t.Fatalf("FAIL: %s: DecodeTlvPackage(%v) =\n\t   %v\n, want %v.", test.description, test.expected, tmpPkt, test.input)
		}
		log.Printf("PASS: DecodeTlvPackage : %s\n", test.description)
	}
}

func TestFromFileToBytesChannel(t *testing.T) {
	dataBinPath := filepath.Join("test-fixtures", "dataTlv.bin")
	tmpBytesChunksToProcess := make(chan []byte, 2)
	go FromFileToBytesChannel(dataBinPath, &tmpBytesChunksToProcess)

	i := 0
	for bytesChunk := range tmpBytesChunksToProcess {
		if !reflect.DeepEqual(bytesChunk, testCasesTlvconverter[i].expected) {
			t.Errorf("FAIL: got %v\n, want %v.", bytesChunk, testCasesTlvconverter[i].expected)
		}
		i++
	}
}

func TestFromBytesChanToMonitoredQueueMultiThread(t *testing.T) {
	// initialization
	dataBinPath := filepath.Join("test-fixtures", "dataTlv.bin")
	tmpBytesChunksToProcess := make(chan []byte, 2)
	go FromFileToBytesChannel(dataBinPath, &tmpBytesChunksToProcess)

	nMq := NewMonitoredQueue(1)
	go FromBytesChanToMonitoredQueueMultiThread(&tmpBytesChunksToProcess, &nMq)
	// checking if the requests has been returned in the same order
	i := 0
	for pkt, ok := nMq.GetFirst(); ok; pkt, ok = nMq.GetFirst() {
		if !reflect.DeepEqual(pkt, testCasesTlvconverter[i].input) {
			t.Fatalf("FAIL: got %v\n, want %v.", pkt, testCasesTlvconverter[i].input)
		}
		i++
	}
}

func TestWriteToCSV(t *testing.T) {
	// initialization
	testMonitoredQueue := NewMonitoredQueue(3)
	testMonitoredQueue.Add(PacketType{A: 1})
	testMonitoredQueue.Add(PacketType{A: 2})
	testMonitoredQueue.Close()
	testCsvPath := filepath.Join("test-fixtures", "testWriteToCSV.csv")
	errchan := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go WriteToCSV(&testMonitoredQueue, testCsvPath, &errchan)
	// wating for the errors
	for err := range errchan {
		t.Error(err.Error())
	}
	//open and check the content
	expectedRead := "1,0,0,0,0,false\n2,0,0,0,0,false\n"
	dat, err := ioutil.ReadFile(testCsvPath)
	if err != nil {
		t.Error(err.Error())
	}
	if !reflect.DeepEqual(string(dat), expectedRead) {
		t.Fatalf("FAIL:  TestWriteToCSV   %v\n, want %v.", string(dat), expectedRead)
	}
	//cleanup
	err = os.Remove(testCsvPath)
	if err != nil {
		t.Error("cleanup: " + err.Error())
	}
}
