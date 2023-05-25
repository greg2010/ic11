package regassign

import "github.com/greg2010/ic11c/internal/ic11/ir"

// RegisterAssigner is an interface that any register assigner type must implement
type RegisterAssigner interface {
	GetRegister(varName ir.IRVar) int
}
