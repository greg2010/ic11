package ic11

import "hash/crc32"

var table = crc32.MakeTable(crc32.IEEE)

func ComputeHash(str string) int {
	checksum := crc32.Checksum([]byte(str), table)
	// Stationeers hash values are represented by 32 bit signed integer, so we force the conversion
	return int(int32(checksum))
}
