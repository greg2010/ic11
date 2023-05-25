package assembler

type testRegisterAssigner struct {
	assignMap map[string]int
}

func (tra *testRegisterAssigner) GetRegister(regName string) int {
	return tra.assignMap[regName]
}
