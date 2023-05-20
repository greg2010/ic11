package ic11

import "hash/crc32"

func ComputeHash(str string) int {
	table := crc32.MakeTable(crc32.IEEE)
	checksum := crc32.Checksum([]byte(str), table)
	return int(int32(checksum))
}
