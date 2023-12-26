package utils

import "sync"

// FuncUponAppExit is a function that will be executed upon app exit.
// Panics will be ignored.
type FuncUponAppExit func()

// IAppExitHelper is a helper to register and execute functions upon app exit.
type IAppExitHelper interface {
	// RegisterFuncUponAppExit registers a function to be executed upon app exit.
	RegisterFuncUponAppExit(funcUponAppExit FuncUponAppExit)

	// ExecuteFunctionsUponAppExit executes all registered functions (LIFO) upon app exit.
	// Panics will be ignored and methods will be executed only once.
	ExecuteFunctionsUponAppExit()
}

// AppExitHelper is a helper to register and execute functions upon app exit.
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

// RegisterFuncUponAppExit registers a function to be executed upon app exit.
func (a *appExitHelper) RegisterFuncUponAppExit(funcUponAppExit FuncUponAppExit) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.funcUponAppExit = append(a.funcUponAppExit, funcUponAppExit)
}

// ExecuteFunctionsUponAppExit executes all registered functions (LIFO) upon app exit.
// Panics will be ignored and methods will be executed only once.
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

	if a.executedFunctionsUponAppExit || len(a.funcUponAppExit) < 1 {
		return
	}

	a.executedFunctionsUponAppExit = true

	for i := len(a.funcUponAppExit) - 1; i >= 0; i-- {
		func(f FuncUponAppExit) {
			defer func() {
				err := recover()
				if err != nil {
					// ignore
				}
			}()

			f()
		}(a.funcUponAppExit[i])
	}
}
