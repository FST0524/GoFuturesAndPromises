package main

import (
	. "../application"
	"fmt"
	"io"
	"net/http"
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

func getRequest(website string) ImplicitPromise {
	return MakeImplicitPromise(func() SharedState {
		resp, error := http.Get(website)
		if error == nil {
			return SharedState{resp, Kept}
		} else{
			return SharedState{error, Broken}
		}

	})
}

func readIO(resp *http.Response) ImplicitPromise {
	return MakeImplicitPromise(func() SharedState {
		bytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body.Close()
			return SharedState{bytes, Kept}
		} else{
			return SharedState{err, Broken}
		}

	})
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

//Testing Implicit Version
//Create a HTTP Get request and print the header
func TestExampleWebsite(t *testing.T) {

	fmt.Print("Sending a HTTP Get Request to a website \n")
	implicitPromise := getRequest("https://www.golem.de/")
	fmt.Print("Get Result \n")
	future := implicitPromise.GetFuture()
	result := future.GetValue()
	switch result.State {
	case Kept:
		request := result.Value.(*http.Response)
		fmt.Printf("Request Header: %s",request.Header)
	case Broken:
		t.Errorf("Promise was broken, the calculation failed with an error: %s \n",result.Value)
		fmt.Print(result.Value)
	default:
		t.Failed()
	}
}

//Failed Version
func TestExampleWebsiteFailed(t *testing.T) {

	fmt.Print("Sending a HTTP Get Request to a website \n")
	implicitPromise := getRequest("https://www.golemFailed12312312123.de/")
	fmt.Print("Get Result \n")
	future := implicitPromise.GetFuture()
	result := future.GetValue()
	switch result.State {
	case Kept:
		t.Error("Promise was kept, but that shouldn't be the case")
	case Broken:
		fmt.Printf("Expected: Promise was broken, the calculation failed with an error: %s \n",result.Value)
		fmt.Print(result.Value)
	default:
		t.Failed()
	}
}

//combined futures and promises
func TestWebAndIO(t *testing.T) {
	fmt.Print("Sending a HTTP Get Request to a website \n")
	HTTPPromise := getRequest("https://www.youtube.com/")
	fmt.Print("Get Result \n")
	future := HTTPPromise.GetFuture()
	result := future.GetValue()
	switch result.State {
	case Kept:
		fmt.Print("Promise was kept therefore processing the response will be processed\n")
		request := result.Value.(*http.Response)
		ioPromise := readIO(request)
		sharedStateIO := ioPromise.GetFuture().GetValue()
		switch sharedStateIO.State {
		case Kept:
			fmt.Print(sharedStateIO.Value)
		case Broken:
			t.Errorf("IOPromise was broken, the calculation failed with an error: %s \n",sharedStateIO.Value)
		}
		fmt.Printf("Request Header: %s",request.Header)
	case Broken:
		fmt.Print(result.Value)
		t.Errorf("HTTPPromise was broken, the calculation failed with an error: %s \n",result.Value)
	default:
		t.Failed()
		break
	}
}
