package ic11

import "fmt"

// MIPS instructions
const (
	move = "move"
	add  = "add"
	sub  = "sub"
	mul  = "mul"
	div  = "div"
	sge  = "sge"
	sgt  = "sgt"
	sle  = "sle"
	slt  = "slt"
	seq  = "seq"
	sne  = "sne"
	j    = "j"
	bnez = "bnez"
	beqz = "beqz"
	sin  = "sin"
	cos  = "cos"
	tan  = "tan"
)

type instruction interface {
	ToString() string
}

type label struct {
	label string
}

func (l *label) ToString() string {
	return fmt.Sprintf("%s:", l.label)
}

type instruction1 struct {
	cmd string
	a   *data
}

func (i1 *instruction1) ToString() string {
	return fmt.Sprintf("%s %s", i1.cmd, i1.a.string())
}

type instruction2 struct {
	cmd  string
	a, b *data
}

func (i2 *instruction2) ToString() string {
	return fmt.Sprintf("%s %s %s", i2.cmd, i2.a.string(), i2.b.string())
}

type instruction3 struct {
	cmd     string
	a, b, c *data
}

func (i3 *instruction3) ToString() string {
	return fmt.Sprintf("%s %s %s %s", i3.cmd, i3.a.string(), i3.b.string(), i3.c.string())
}

//nolint:unused
type instruction4 struct {
	cmd        string
	a, b, c, d *data
}

//nolint:unused
func (i4 *instruction4) ToString() string {
	return fmt.Sprintf("%s %s %s %s %s", i4.cmd, i4.a.string(), i4.b.string(), i4.c.string(), i4.d.string())
}

type data struct {
	register *register
	value    float64
	label    string
}

func (d *data) string() string {
	if d.register != nil {
		return d.register.name()
	}

	if d.label != "" {
		return d.label
	}

	return fmt.Sprintf("%v", d.value)
}
func (d *data) isRegister() bool {
	return d.register != nil
}

func (d *data) isTemporaryRegister() bool {
	return d.register != nil && d.register.temporary
}

func (d *data) isValue() bool {
	return d.register == nil && d.label == ""
}

func newNumData(v float64) *data {
	return &data{value: v}
}

func newRegisterData(register *register) *data {
	return &data{register: register}
}

func newLabelData(label string) *data {
	return &data{label: label}
}

type register struct {
	i         int
	allocated bool
	temporary bool
}

func newRegister(i int, temporary bool) *register {
	return &register{i: i, allocated: false, temporary: temporary}
}

func (r *register) allocate() {
	r.allocated = true
}

func (r *register) deallocate() {
	r.allocated = false
}

func (r *register) name() string {
	return fmt.Sprintf("r%d", r.i)
}
