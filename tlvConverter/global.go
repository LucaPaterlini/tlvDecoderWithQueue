package tlvconverter

const (
	// PACKET is the PacketType marker
	PACKET = 1
	// UINT64 is the unsigned integer at 64 marker
	UINT64 = 2
	// FLOAT64 is the unsigned float at 64 marker
	FLOAT64 = 3
	// BOOL is the bool marker
	BOOL = 4
)

// PacketType is the struct implementing the packet saved as tlv byte array in the input array.
type PacketType struct {
	A uint64
	B float64
	C float64
	D float64
	E float64
	F bool
}
