package ic11

import (
	"fmt"
	"strconv"
	"strings"
)

// MIPS instructions
const (
	move  = "move"
	add   = "add"
	sub   = "sub"
	mul   = "mul"
	div   = "div"
	and   = "and"
	or    = "or"
	xor   = "xor"
	nor   = "nor"
	sge   = "sge"
	sgt   = "sgt"
	sle   = "sle"
	slt   = "slt"
	seq   = "seq"
	sne   = "sne"
	j     = "j"
	bnez  = "bnez"
	beqz  = "beqz"
	sin   = "sin"
	cos   = "cos"
	tan   = "tan"
	mod   = "mod"
	l     = "l"
	lb    = "lb"
	lr    = "lr"
	ls    = "ls"
	s     = "s"
	sb    = "sb"
	yield = "yield"
	bge   = "bge"
	bgt   = "bgt"
	ble   = "ble"
	blt   = "blt"
	beq   = "beq"
	bne   = "bne"
	abs   = "abs"
	acos  = "acos"
	asin  = "asin"
	atan  = "atan"
	ceil  = "ceil"
	exp   = "exp"
	floor = "floor"
	log   = "log"
	max   = "max"
	min   = "min"
	sqrt  = "sqrt"
	round = "round"
	trunc = "trunc"
	rand  = "rand"
)

func replaceLabelData(lblMap map[string]int, d *data) (*data, error) {
	if d.isLabel() {
		addr, found := lblMap[d.label]
		if !found {
			return nil, ErrUnknownLabel
		}

		return newNumData(float64(addr)), nil
	}

	return d, nil
}

type instruction interface {
	ToString() string
	ReplaceLabels(lblMap map[string]int) error
}

type label struct {
	label string
}

func (l *label) ToString() string {
	return fmt.Sprintf("%s:", l.label)
}

func (l *label) ReplaceLabels(lblMap map[string]int) error {
	return nil
}

type instructionN struct {
	cmd  string
	args []*data
}

func newInstructionN(cmd string, args ...*data) *instructionN {
	return &instructionN{cmd, args}
}

func (in *instructionN) ToString() string {
	if len(in.args) == 0 {
		return in.cmd
	}
	strArgs := []string{}
	for _, arg := range in.args {
		strArgs = append(strArgs, arg.string())
	}

	return fmt.Sprintf("%s %s", in.cmd, strings.Join(strArgs, " "))
}

func (in *instructionN) ReplaceLabels(lblMap map[string]int) error {
	for i := range in.args {
		var err error
		in.args[i], err = replaceLabelData(lblMap, in.args[i])
		if err != nil {
			return err
		}
	}

	return nil
}

type dataType int

const (
	lbl dataType = iota
	rgstr
	fltVal
	dvc
	hshStr
	str
)

type data struct {
	typ        dataType
	register   *register
	device     string
	floatValue float64
	label      string
	hashString string
	strValue   string
}

func (d *data) string() string {
	switch d.typ {
	case lbl:
		return d.label
	case rgstr:
		return d.register.name()
	case fltVal:
		return strconv.FormatFloat(d.floatValue, 'f', -1, 64)
	case dvc:
		return d.device
	case hshStr:
		return fmt.Sprintf("HASH(\"%s\")", d.hashString)
	case str:
		return d.strValue
	default:
		// Should not ever happen
		return ""
	}
}

func (d *data) isLabel() bool {
	return d.typ == lbl
}

func (d *data) isRegister() bool {
	return d.typ == rgstr
}

func (d *data) isDevice() bool {
	return d.typ == dvc
}

func (d *data) isTemporaryRegister() bool {
	return d.typ == rgstr && d.register.temporary
}

func (d *data) isFloatValue() bool {
	return d.typ == fltVal
}

func newNumData(v float64) *data {
	return &data{typ: fltVal, floatValue: v}
}

func newRegisterData(register *register) *data {
	return &data{typ: rgstr, register: register}
}

func newLabelData(label string) *data {
	return &data{typ: lbl, label: label}
}

func newHashData(hashStr string) *data {
	return &data{typ: hshStr, hashString: hashStr}
}

func newStringData(strVal string) *data {
	return &data{typ: str, strValue: strVal}
}

func newDeviceData(device string) *data {
	return &data{typ: dvc, device: device}
}

type register struct {
	id        int
	allocated bool
	temporary bool
}

func newRegister(id int, temporary bool) *register {
	return &register{id: id, allocated: false, temporary: temporary}
}

func (r *register) allocate() {
	r.allocated = true
}

func (r *register) deallocate() {
	r.allocated = false
}

func (r *register) name() string {
	return fmt.Sprintf("r%d", r.id)
}
