package future

import "time"

// Tasks
//_______________________________________________________________________________________________

/*
 -Task-: Schritt: Grundlegende Implementierung (Done)
 -Task-: Schritt: Beispiel schreiben, um Funktionalität zu testen (Done)
 -ToDo-: Schritt: Neue Funktionalitäten hinzufügen und Optimieren
 ToDo: Schritt: " testen
*/

// Implementation
//_______________________________________________________________________________________________
/*
 SharedState is used to later receive the current state of the future as well as the potentally computet value
 Variables:
	STATE has three values: 0 = Completed, 1 = Waiting, 2 = Broken //How is this possible to receive
	VALUE keeps the result that will be available
*/
const (
	Kept    = iota // Completed
	Broken  = iota // Failed
	Timeout = iota // Failed
	//Processing = iota //Still calculating the result --- Really needed? What are the benefits?
)

type SharedState struct {
	Value interface{}
	State int8
}

// Future
// _______________________________________________________________________________________________

/*
 -Task-: Annahmen
 -Task-: Only store one value and read-only
 -Solution-: Create unbuffered channel
*/
/*
 Create Future as placeholder object the value will be eventually available
*/
type Future chan SharedState

// Same Problem with the user will receive a SharedState

func (future Future) GetValue() SharedState {
	sharedState := <-future
	return sharedState
}

// -Experimental Function-: Add select statement and add timeout to stop
// -Info- Due to the change in the previous implementation
func (future Future) GetValueWithTimeout(seconds int) SharedState {
	var sharedState SharedState
	select {
	case sharedState = <-future:
		return sharedState
	case <-time.After(time.Duration(seconds) * time.Second):
		sharedState = SharedState{Value: 0, State: Timeout}
		return sharedState
	}
}

// Future - Optional Extension
// _______________________________________________________________________________________________

func (future Future) OnPromiseBroken(work func()) {
	result := future.GetValue()
	if result.State == Broken {
		work()
	}
}

func (future Future) OnPromiseKept(work func(i interface{})) {
	result := future.GetValue()
	if result.State == Kept {
		work(result.Value)
	}
}

// Promise
//_______________________________________________________________________________________________
/*
 -Task-: Annahmen
 -Task-: 1 Promise erstellt den Future
 -Task-: 2 Promise implizit
 -Task-: 2 Promise explizit
 -Task-: 3 Promise hat die Funktion getFuture um eine Instanz des Futures zu setzen
 -Task-: 4 Promise hat eine Funktion setValue, um den Wert des Futures zu setzen, dadurch wird er automatisch in den Zustand completed gesetzt und ist erfüllt
*/
type Promise struct {
	LinkedFuture Future
}
type ExplicitPromise Promise
type ImplicitPromise Promise

// Info: Promise in this case is just the calculation and processing for the value
// Generates a new Promise while directly starting the calculation
// (1),(2) will be solved
// Input: A task which returns a SharedState
// Output: Return the created Promise
// Effect: the calculation starts immediately
// Implicit means that the calculation will be starting without a trigger
func MakeImplicitPromise(calcFunction func() SharedState) ImplicitPromise {
	//Closure used here
	promise := ImplicitPromise{LinkedFuture: make(Future)}
	go func() {
		// ??? This will force the users to return a Shared. User friendly?? !!!
		ret := calcFunction()
		// -Task-: Multiple Read Request for the Future
		// -Solution-: Keep sending the result to the channel
		// -Problem-: Resources are wasted? No because the channel should be blocking after each send
		// therefore it will wait until the getFuture function is called
		for {
			promise.LinkedFuture <- ret
		}
	}()
	return promise
}

// Get Future(ReturnValue) from Promise(Base)
// Effects: Will return the future for the promise
func (promise ImplicitPromise) GetFuture() Future {
	return promise.LinkedFuture
}

// Info: Classic C++ Future and Promise Functionality
// Generates a new Future (nothing else)
// Explicit means that the calculation will not be started. A trigger or a promise.SetValue is needed.
func MakeExplicitPromise() ExplicitPromise {
	promise := ExplicitPromise{LinkedFuture: make(Future)}
	return promise
}

/*
Reason: On Promise(Base) get the Future to set the value
Effects: Will set the status to completed and the value can be received through Future.getValue()
-Task- Potential Problem: What happened if the function as well as implizit executes at the same time => Deadlock not on another thread
-Fix-: Create a implicit and an explicit Promise (Because it's not allowed by our definition
*/
func (promise ExplicitPromise) PromiseValue(input interface{}) {
	//Blocking until value is send
	promise.LinkedFuture <- SharedState{input, Kept}
	//However the value should be read more than one time therefore we need to send the result more than one
	//time to the channel. To prevent the blocking we need to use a goroutine
	go func() {
		for {
			promise.LinkedFuture <- SharedState{input, Kept}
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
// Possible Extensions - Add Timeout to calculation
//                     - Add a cancel function
// 					   - Get State to check if still processing
//
//-> However is problematic because we don't know how expensive
//   the calculation is -> So make it optional and let the user set the amount of time
