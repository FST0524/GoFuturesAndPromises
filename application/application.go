// Github: https://github.com/FST0524/GoFuturesAndPromises

package future

import (
	"fmt"
	"time"
)

// Types
//_______________________________________________________________________________________________

type FutureResult struct {
	ValueOrError interface{}
	IsError bool
}

type TimeoutError struct{}
type PromiseError struct{
	Reason string
}

func (e TimeoutError) Error() string {
	return "Timeout error occurred.\n"
}

func (e PromiseError) Error() string {
	return fmt.Sprintf("The promise was broken.\n Reason: %s ",e.Reason)
}

// Future
//-Task-: Only store one value and read-only
//-Solution-: Limit buffer e.g. unbuffered
// _______________________________________________________________________________________________

//Create a Future as a placeholder object for the value that will be eventually available
type Future chan FutureResult

//GetResult() - Retrieve the value of the future
//Properties: Blocking
func (future Future) GetResult() (interface{},error) {
	futResult := <-future
	if futResult.IsError == false {
		return futResult.ValueOrError, nil
	} else {
		return nil, futResult.ValueOrError.(error)
	}
}


// Future - Possible Extensions
// _______________________________________________________________________________________________

//Retrieve the value of the future with a timeout
//If the value can't be received the function will return a TimeoutError
//Properties: Partially Blocking (until timeout occures)
func (future Future) GetResultWithTimeout(seconds int) (interface{},error)  {
	select {
	case futResult := <-future:
		if futResult.IsError == false {
			return futResult.ValueOrError, nil
		} else {
			return nil, futResult.ValueOrError.(error)
		}
	case <-time.After(time.Duration(seconds) * time.Second):
		return nil, TimeoutError{}
	}
}

//If the calculation fails execute a given function
//Properties: Blocking
//Reason: Try to capsulate most of the functionality to make it easier to use
func (future Future) OnPromiseBroken(execFunc func()) {
	_, err := future.GetResult()
	if err != nil {
		execFunc()
	}
}

//Process the result with a given function
//Properties: Blocking
//Reason: Try to capsulate most of the functionality to make it easier to use
func (future Future) OnPromiseKept(execFunc func(i interface{})) {
	res, err := future.GetResult()
	if err == nil {
		execFunc(res)
	}
}

//Process the result with a given function
//If the value couldn't be received then execute errorFunction
//Properties: Blocking
//Reason: Try to capsulate most of the functionality to make it easier to use
func (future Future) OnResolvedWithTimeout(errorFunc func(),successFunc func(i interface{}),seconds int) {
	res, err := future.GetResultWithTimeout(seconds)
	if err == nil {
		successFunc(res)
	} else {
		errorFunc()
	}
}

// Promise
//_______________________________________________________________________________________________

// Keep every promise bound to their own future
type Promise struct {
	LinkedFuture Future
}

// Promise
//_______________________________________________________________________________________________

type ImplicitPromise Promise

// Info: Promise in this case is just the calculation and processing for the value
// Generates a new Promise while directly starting the calculation
// Input: A task which returns a FutureResult
// Output: Return the created Promise
// Effect: the calculation starts immediately
// Implicit in this case means the calculation will start without a specific trigger
func MakeImplicitPromise(calcFunction func() FutureResult) ImplicitPromise {
	//Closure used here
	//It's possible to make the func easier with directly returning the future
	promise := ImplicitPromise{LinkedFuture: make(Future)}
	go func() {
		ret := calcFunction()
		// -Task-: Multiple Read Request for the Future
		// -Solution-: Keep sending the result to the channel
		// -Problem-: Resources are wasted? No because the channel should be blocking after each send
		// therefore it will wait until the getFuture function is called
		// But this could be causing a memory leak!!! -> Solution Read one time
		for {
			promise.LinkedFuture <- ret
		}
	}()
	return promise
}

// Get the associated Future to a ImplicitPromise
func (promise ImplicitPromise) GetFuture() Future {
	return promise.LinkedFuture
}

// Promise - Functionality similar to C++ Boost Future and Promise
//_______________________________________________________________________________________________

type ExplicitPromise Promise

// Creates the association between the ExplicitPromise and the Future
// Explicit means that the calculation will not be started.
// A trigger is needed. The corresponding trigger is the function PromiseValue
// In Go this implementation may not be need because the same effect could be easily achieved with only channels
func MakeExplicitPromise() ExplicitPromise {
	promise := ExplicitPromise{LinkedFuture: make(Future)}
	return promise
}

/*
On Promise(Base) get the Future to set the value
Will set the status to completed and the value can be received through Future.getResult()
Important: If a ExplicitPromise is created the PromiseValue has to be called because else the promise will stay in blocked state
*/
func (promise ExplicitPromise) PromiseValue(input interface{}) {
	//Blocking until value is send
	promise.LinkedFuture <- FutureResult{input, false}
	//However the value should be read more than one time therefore we need to send the result more than one
	//time to the channel. To prevent the blocking we need to use a goroutine
	//Possible Problem: Memory Leak
	go func() {
		for {
			promise.LinkedFuture <- FutureResult{input, false}
		}
	}()
}

// Return the future for the promise
func (promise ExplicitPromise) GetFuture() Future {
	return promise.LinkedFuture
}

// Promise - Optional Extension
// _______________________________________________________________________________________________
// Possible Extensions
//                     - Add a cancel function using context (Recommended)
// 		       - Get State to check if still processing (Test case) (Optional)

