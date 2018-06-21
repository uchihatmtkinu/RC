package rccache

import (
	"testing"

	"github.com/uchihatmtkinu/RC/account"
)

func TestOutToData(t *testing.T) {
	var acc1, acc2 account.RcAcc
	acc1.New("1")
	acc2.New("2")
	var test1, test2 DbRef
	test1.New(1, acc1.Pri)
	test2.New(2, acc2.Pri)
	t.Error(`Check`)
}
