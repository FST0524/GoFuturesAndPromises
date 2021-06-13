package main

import (
	. "../application"
	"fmt"
	"testing"
	"time"
)

//_______________________________________________________________________________________________
// Helper functions
//_______________________________________________________________________________________________
func printState(state SharedState) {
	switch state.State {
	case Kept:
		fmt.Printf("Future completed. The value was %d \n", state.Value)
	case Broken:
		fmt.Print("Future was broken.\n")
	default:
		fmt.Print("Error: State was corrupted.\n")
	}
}

func testImplicitExecution(testValue int) ImplicitPromise {
	return MakeImplicitPromise(func() SharedState {
		if testValue < 10 {
			//Simulating long working process
			time.Sleep(10 * time.Second)
			return SharedState{Value: testValue, State: Kept}
		} else {
			return SharedState{Value: testValue, State: Broken}
		}
	})
}

func testPrintValue(f Future, ch chan interface{}) {
	x := f.GetValue().Value
	ch <- x
	fmt.Printf("TestPrintValue: %d \n", x)
}

func testCalculationValue(p ExplicitPromise) {
	x := 0
	for i := 0; i < 10; i++ {
		x = x + i
	}
	p.PromiseValue(x)
}

//_______________________________________________________________________________________________
// Tests
//_______________________________________________________________________________________________

//Testing Implicit Version
// Default calculation of a given function in another thread
func TestImplicitImplementation(t *testing.T) {
	fmt.Print("Trying to get Promise \n")
	pro := testImplicitExecution(5)
	fmt.Print("Trying to get Future \n")
	fut := pro.GetFuture()

	fmt.Print("Trying to print result\n")
	x := fut.GetValue()
	printState(x)
	if x.Value != 5 {
		t.Errorf("FAILED - Wrong value %d was received but should have been 5", x.Value)
	}
}

//Testing Explicit Version
//Promise a Value and wait for it in another thread
func TestExplicitWithPromising(t *testing.T) {
	fmt.Print("Creating Promise and Future\n")
	pro := MakeExplicitPromise()
	fmt.Print("Starting go function to wait for future value \n")
	fut := pro.GetFuture()

	ch := make(chan interface{})
	go testPrintValue(fut, ch)
	fmt.Print("Promising Value\n")
	pro.PromiseValue(10)
	time.Sleep(time.Second * 5)
	x := <-ch
	if x != 10 {
		t.Errorf("FAILED - Wrong value %d was received but should have been 10", x)
	}
}

//Testing Explicit Version
//Promise a value in another thread and wait for it in main thread
func TestExplicitWithPromisingInThread(t *testing.T) {
	fmt.Print("Creating Promise and Future\n")
	pro := MakeExplicitPromise()
	fmt.Print("Starting go function to calculate the value\n")
	go testCalculationValue(pro)
	fut := pro.GetFuture()
	fmt.Print("Trying to get Future Value \n")
	x := fut.GetValue()
	printState(x)
	if x.Value != 45 {
		t.Errorf("FAILED - Wrong value %d was received but should have been 45", x.Value)
	}
}
