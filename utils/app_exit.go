package utils

import "sync"

type FuncUponAppExit func()
type IAppExitHelper interface {
	RegisterFuncUponAppExit(funcUponAppExit FuncUponAppExit)
	ExecuteFunctionsUponAppExit()
}

var AppExitHelper IAppExitHelper = &appExitHelper{
	mutex:                        &sync.Mutex{},
	executedFunctionsUponAppExit: false,
	funcUponAppExit:              nil,
}

type appExitHelper struct {
	mutex                        *sync.Mutex
	executedFunctionsUponAppExit bool
	funcUponAppExit              []FuncUponAppExit
}

func (a *appExitHelper) RegisterFuncUponAppExit(funcUponAppExit FuncUponAppExit) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.funcUponAppExit = append(a.funcUponAppExit, funcUponAppExit)
}

func (a *appExitHelper) ExecuteFunctionsUponAppExit() {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	r := recover()

	defer func() {
		if r != nil {
			PrintlnStdErr("ERR: rethrow panic")
			panic(r)
		}
	}()

	if a.executedFunctionsUponAppExit {
		return
	}

	a.executedFunctionsUponAppExit = true

	for _, f := range a.funcUponAppExit {
		func(f FuncUponAppExit) {
			defer func() {
				err := recover()
				if err != nil {
					// ignore
				}
			}()

			f()
		}(f)
	}
}
