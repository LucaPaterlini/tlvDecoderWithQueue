package tlvconverter

import "errors"

var testCasesTlvconverter = []struct {
	description string
	input       PacketType
	expected    []byte
}{
	{
		description: "empty ",
		input:       PacketType{},
		expected:    []byte{1, 18, 2, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4, 1, 0},
	},
	{
		description: "only a ",
		input:       PacketType{A: 12},
		expected:    []byte{1, 18, 2, 1, 12, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4, 1, 0},
	},
	{
		description: "only b",
		input:       PacketType{B: 0.1},
		expected: []byte{1, 25, 2, 1, 0, 3, 8, 63, 185, 153, 153, 153, 153, 153, 154, 3, 1,
			0, 3, 1, 0, 3, 1, 0, 4, 1, 0},
	},
	{
		description: "only f that is boolean",
		input:       PacketType{F: true},
		expected:    []byte{1, 18, 2, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4, 1, 1},
	},
	{
		description: "full package",
		input:       PacketType{A: 1, B: 1.1, C: 2.2, D: 3.3, E: 4.4, F: true},
		expected: []byte{1, 46, 2, 1, 1, 3, 8, 63, 241, 153, 153, 153, 153, 153, 154, 3, 8, 64, 1, 153,
			153, 153, 153, 153, 154, 3, 8, 64, 10, 102, 102, 102, 102, 102, 102, 3, 8, 64, 17, 153, 153, 153, 153, 153,
			154, 4, 1, 1},
	},
}

// testCasesVltconverter
var testCasesTlvconverterFailure = []struct {
	description   string
	input         []byte
	expectedError error
}{
	{
		description:   "testing the wrong value for the package marker",
		input:         []byte{0, 18, 2, 1, 12, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4, 1, 0},
		expectedError: errors.New("not a PacketType tlv bytes representation"),
	},
	{
		description:   "testing wrong len package",
		input:         []byte{1, 18, 2, 1, 12, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4},
		expectedError: errors.New("package corrupted, too short"),
	},
	{
		description:   "corrupted triplet",
		input:         []byte{1, 17, 2, 1, 12, 3, 1, 0, 3, 1, 0, 3, 1, 0, 3, 1, 0, 4, 1},
		expectedError: errors.New("package corrupted, too short"),
	},
}
