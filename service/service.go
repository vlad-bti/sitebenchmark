package service

import (
	"errors"
	"sync"
)

type ServiceManager struct {
	benchmarkCenter BenchmarkCenter
}

var (
	smOnce                sync.Once
	instance              *ServiceManager
	errServiceDuplicated  = errors.New("duplicated service")
	errServiceUnsupported = errors.New("unsupported service")
)

func Mgr() *ServiceManager {
	smOnce.Do(func() {
		instance = &ServiceManager{}
	})
	return instance
}

func (c *ServiceManager) BenchmarkCenterService() BenchmarkCenter {
	return c.benchmarkCenter
}

func (c *ServiceManager) Register(svcs ...interface{}) error {
	for _, svc := range svcs {
		switch s := svc.(type) {
		case BenchmarkCenter:
			if c.benchmarkCenter != nil {
				return errServiceDuplicated
			}
			c.benchmarkCenter = s
		default:
			return errServiceUnsupported
		}
	}
	return nil
}
