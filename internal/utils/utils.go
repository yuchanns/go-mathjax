package utils

import (
	"strings"
	"sync"
)

type ErrGroup struct {
	locker *sync.Mutex
	errs   []error
	hasErr bool
}

func NewErrGroup() *ErrGroup {
	return &ErrGroup{
		locker: &sync.Mutex{},
		errs:   make([]error, 20),
		hasErr: false,
	}
}

func (group *ErrGroup) Append(err error) {
	group.locker.Lock()
	defer group.locker.Unlock()
	if !group.hasErr {
		group.hasErr = true
	}
	group.errs = append(group.errs, err)
}

func (group *ErrGroup) HasErr() bool {
	return group.hasErr
}

func (group *ErrGroup) Error() string {
	errStrs := make([]string, 0, len(group.errs))
	for _, err := range group.errs {
		errStrs = append(errStrs, err.Error())
	}
	return strings.Join(errStrs, ";")
}
