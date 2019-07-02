#  Tlv Decoder With Queue

## Intro

This package has been written as part of a showcase of my knowledge of golang, It's not intended to be production ready and in the last part I have proposed a two more reality resilient solutions.

I have used [goland](https://www.jetbrains.com/go/) as IDE.

I have used [exalys](https://github.com/exercism/exalysis) to run the suite of automated go tools.

I have made the following assumptions :
- The encoding type takes 1 byte, the length 1 byte, the value up to 8 bytes (float64 or uint64).
- The encoding is Big Endian
- The length represents the length in bytes of the value even for nested fields.

# Problem

Given a binary file and its Tlv encoding, produce a CSV with its content,
in a human-readable format.

To make the solution more scalable over multiple cores create a queue and to make it more maintainable
monitor its r/w rate and its state of filling.

## Implemented Solution

Standalone program, with the name of the input and output file as parameters,
and as optional flags the size of the queue, the delay in between each monitor function
calls and the delay for the reads.


The last flag is useful for testing to add a delay in the read to check how
the handler of the delay on the write adapts in case of slow read from the queue.

### Execution:

A single goroutine read from the binary file and split it in chunks of bytes
and then put each chunk on a queue.

A second goroutine starts a pool of goroutines, adding a delay in between them dictated
by the write delay parameter in the monitored queue, that decode the chunks of bytes and put
the results of each one in its receiver channel, the launcher routine loop sequentially
over the receiver channels and add the results to the monitored queue.

There is a drawback in this approach witch goal is to use multithreading
and keep the order of the packets, if the end of the goroutines
in the pool is not guaranteed the launcher routine could get stuck
until all its launched goroutines end.

This is not the case because the end of the decoding, Tlv bytes chunks to packet function is guarantee in O(1) time.

Meanwhile a goroutine to monitor the status of the loading has been started and
at every x seconds, where x is the number received from the flag for the delay
of the monitor calls, a call to the monitor function is invoked showing
the current status of the filling of the map and the rate of the packet in and out
since the last call.

There is a last goroutine to mention that is the one that converts the packets
received throw the monitored queue and write them as records on the CSV
received as parameter.

## Usage

`
go build main.go [-dr=0 -secload=1 -bufsize=3] "pathInputFile" "pathOutputCsv"
`

## Real World implementations

This example is clearly oversimple, a single machine where threads can communicate accessing the same memory.


### Rabbit Message Queue

A more scalable solution is one that uses MessageQueue like
[RabbitMq](https://www.rabbitmq.com/) instead of the struct Monitored Queue used in this implementation.

It gave us the primitives on the status of the filling of the queue,
on its read and write rates and allow us to have multiple sources
of data that read at the same time the same sequence of data
(eg. redundant sensors).

It allows as well to differentiate over the packet to read and produce
multiple CSV tailored for the sensor to read.
(eg 4 sensors, 1 server with RabbitMq, 1 server that writes a CSV
with the packets of sensor 1 and 2, 1 server that writes a CSV with
all the queue etc).

### Using Stream gRpc as MessageQueue

Another interesting approach that requires a bit of extra work is using
[grpc stream](https://grpc.io/docs/tutorials/basic/go/) opening a communication
stream over a gRPC in between the device that parses the binary file
and the server that write the csv, keeping the same primitives
of this solution regarding writing a reading from the queue
and substituting only the implementation of the monitored queue.

To have the monitor functionality both devices have to run the goroutine that
monitor the status of the queue and have the declaration of a struct that
keep track of writes on the sender and read on the reader and sends
the information over a second bidirectional stream in between them.  

