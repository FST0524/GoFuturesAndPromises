package future

import (
	"fmt"
	"time"
)

// Tasks
//_______________________________________________________________________________________________

/*
 -Task-: Schritt: Grundlegende Implementierung (Done)
 -Task-: Schritt: Beispiel schreiben, um Funktionalität zu testen (Done)
 -ToDo-: Schritt: Neue Funktionalitäten hinzufügen und Optimieren
 ToDo: Schritt: " testen
*/

// Types
//_______________________________________________________________________________________________

type FutureResult struct {
	ValueOrError interface{}
	IsError bool
}

type error interface {
	Error() string
}

type TimeoutError struct{

}

func (e TimeoutError) Error() string {
	return fmt.Sprintf("Timeout error occured.")
}

// Future
//-Task-: Only store one value and read-only
//-Solution-: Limit buffer e.g. unbuffered
// _______________________________________________________________________________________________


//Create Future as placeholder object the value will be eventually available
type Future chan FutureResult

//Retrieve the value of the future
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

//Retrieve the value of the future with a specific duration until a timeout occures
//Properties: Blocking (until timeout occures)
func (future Future) GetResultWithTimeout(seconds int) (interface{},error)  {
	select {
	case futResult := <-future:
		if futResult.ValueOrError == false {
			return futResult.ValueOrError, nil
		} else {
			return nil, futResult.ValueOrError.(error)
		}
	case <-time.After(time.Duration(seconds) * time.Second):
		return nil, TimeoutError{}
	}
}

//If the calculation failed execute a given function
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
//Properties: Blocking
//Reason: Try to capsulate most of the functionality to make it easier to use
func (future Future) OnResolvedWithTimeout(errorFunc func(),successFunc func(i interface{}),seconds int) {
	res, err := future.GetResultWithTimeout(seconds)
	if err == nil {
		successFunc(res)
	} else {
		err.Error()
		errorFunc()
	}
}

// Promise
// -Task-: 1 Promise erstellt den Future
// -Task-: 2 Promise implizit
// -Task-: 2 Promise explizit
// -Task-: 3 Promise hat die Funktion getFuture um eine Instanz des Futures zu setzen
// -Task-: 4 Promise hat eine Funktion setValue, um den Wert des Futures zu setzen, dadurch wird er automatisch in den Zustand completed gesetzt und ist erfüllt
//_______________________________________________________________________________________________

// Keep every promise bound to their own promise
type Promise struct {
	LinkedFuture Future
}

// Promise
//_______________________________________________________________________________________________

type ImplicitPromise Promise

// Info: Promise in this case is just the calculation and processing for the value
// Generates a new Promise while directly starting the calculation
// (1),(2) will be solved
// Input: A task which returns a SharedState
// Output: Return the created Promise
// Effect: the calculation starts immediately
// Implicit means that the calculation will be starting without a trigger
func MakeImplicitPromise(calcFunction func() FutureResult) ImplicitPromise {
	//Closure used here
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

// Get Future (ReturnValue) from Promise(Base)
// Effects: Will return the future for the promise
func (promise ImplicitPromise) GetFuture() Future {
	return promise.LinkedFuture
}

// Promise - Functionality is similar to C++ Boost Future and Promise
//_______________________________________________________________________________________________

type ExplicitPromise Promise

// Info:
// Generates a new Future (nothing else)
// Explicit means that the calculation will not be started. A trigger or a promise. SetValue is needed.
// In Go this implementation may not be need because the same effect could be easiyl achieved with only channels
func MakeExplicitPromise() ExplicitPromise {
	promise := ExplicitPromise{LinkedFuture: make(Future)}
	return promise
}

/*
Reason: On Promise(Base) get the Future to set the value
Effects: Will set the status to completed and the value can be received through Future.getValue()
-Task- Potential Problem: What happened if the function as well as implizit executes at the same time => Deadlock not on another thread
-Fix-: Create a implicit and an explicit Promise (Because it's not allowed by our definition)
*/
func (promise ExplicitPromise) PromiseValue(input interface{}) {
	//Blocking until value is send
	promise.LinkedFuture <- FutureResult{input, false}
	//However the value should be read more than one time therefore we need to send the result more than one
	//time to the channel. To prevent the blocking we need to use a goroutine

	//Problematic: Possible Memory Leak
	go func() {
		for {
			promise.LinkedFuture <- FutureResult{input, false}
		}
	}()
}

// Get Future(ReturnValue) from Promise(Base)
// Effects: Will return the future for the promise
func (promise ExplicitPromise) GetFuture() Future {
	return promise.LinkedFuture
}

// Promise - Optional Extension
// _______________________________________________________________________________________________
// Possible Extensions
//                     - Add a cancel function (Macht Sinn) (bin zu dumm dafür) (Kontexte)
// 					   - Get State to check if still processing (Test case)
//
//-> However is problematic because we don't know how expensive
//   the calculation is -> So make it optional and let the user set the amount of time
