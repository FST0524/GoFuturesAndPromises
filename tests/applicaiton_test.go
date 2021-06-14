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
// Tests
//_______________________________________________________________________________________________

//Testing Implicit Version
//Simple default calculation of a given function in another thread
func TestImplicitImplementation(t *testing.T) {

	pro := testImplicitExecution(5)
	fut := pro.GetFuture()
	x, err:= fut.GetResult()

	if err == nil {
		fmt.Printf("Received FutureResult Value: %d \n", x)
	}else{
		err.Error()
	}

	if x != 5 {
		t.Errorf("Wrong value %d was received from FutureResult. Should have been 5\n", x)
	}
}

//Testing Explicit Version
//Promise a Value and wait for it in another thread
func TestExplicitWithPromising(t *testing.T) {
	pro := MakeExplicitPromise()
	fut := pro.GetFuture()
	checkResult := make(chan int)

	go testPrintValue(fut, checkResult)
	pro.PromiseValue(10)

	time.Sleep(time.Second * 5)

	x := <-checkResult
	if x != 10 {
		t.Errorf("FAILED - Wrong value %d was received but should have been 10\n", x)
	}
}

//Testing Explicit Version
//Promise a value in another thread and wait for it in main thread
func TestExplicitWithPromisingInThread(t *testing.T) {
	pro := MakeExplicitPromise()
	go testCalculationValue(pro)
	fut := pro.GetFuture()

	x,err := fut.GetResult()
	if err == nil {
		if x == 45 {
			fmt.Print("Correct Value received.\n")
		}else{
			t.Errorf("Wrong value %d was received but should have been 45.\n", x)
		}
	}else {
		err.Error()
		t.Errorf("Error occured during proccessing.\n")
	}

}

//Testing Implicit Version
//Create a HTTP Get request and print the header
func TestExampleWebsite(t *testing.T) {

	fmt.Print("Sending a HTTP Get Request to a website \n")
	implicitPromise := getRequest("https://www.golem.de/")

	fmt.Print("Get Result \n")
	future := implicitPromise.GetFuture()

	result, err := future.GetResult()
	if err == nil{
		request := result.(*http.Response)
		fmt.Printf("Request Header: %s",request.Header)
	}else{
		err.Error()
	}
}

//Failed Version
func TestExampleWebsiteFailed(t *testing.T) {

	fmt.Print("Sending a HTTP Get Request to a website \n")
	implicitPromise := getRequest("https://www.golemFailed12312312123.de/")
	fmt.Print("Get Result \n")
	future := implicitPromise.GetFuture()
	result, err := future.GetResult()

	if err == nil{
		t.Error("Promise was kept, but that shouldn't be the case")
	}else{
		fmt.Printf("Expected: Promise was broken, the calculation failed with an error: %s \n",result)
	}
}

//combined futures and promises
func TestWebAndIO(t *testing.T) {
	fmt.Print("Sending a HTTP Get Request to a website \n")
	HTTPPromise := getRequest("https://www.youtube.com/")
	fmt.Print("Get Result \n")
	future := HTTPPromise.GetFuture()
	resultWeb, webErr := future.GetResult()

	if webErr == nil {
		fmt.Print("Promise was kept therefore processing the response will be executed\n")
		request := resultWeb.(*http.Response)
		ioPromise := readIO(request)

		resultIO, IOErr := ioPromise.GetFuture().GetResult()
		if IOErr == nil{
			fmt.Print(resultIO)
		} else{
			t.Errorf("IOPromise was broken, the calculation failed with an error: %s \n",resultIO)
		}
	}else{
		fmt.Print(resultWeb)
		t.Errorf("HTTPPromise was broken, the calculation failed with an error: %s \n",resultWeb)
	}
}

// Test the GetResultWithTimeout function
// The timeout wont happen
func TestGetResultWithTimeout(t *testing.T) {

}
// Test the GetResultWithTimeout function
// The timeout will happen
func TestGetResultWithOccurringTimeout(t *testing.T) {

}

// Test the OnResolvedWithTimeout function
// SuccessFunc should be executed
func TestOnResolvedWithTimeoutSuccess(t *testing.T) {

}

// Test the OnResolvedWithTimeout function
// ErrorFunc should be executed
func TestOnResolvedWithTimeoutError(t *testing.T) {

}

//_______________________________________________________________________________________________
// Helper functions
//_______________________________________________________________________________________________

func testImplicitExecution(testValue int) ImplicitPromise {
	return MakeImplicitPromise(func() FutureResult {
		if testValue < 10 {
			//Simulating long working process
			time.Sleep(10 * time.Second)
			return FutureResult{ValueOrError: testValue, IsError: false}
		} else {
			return FutureResult{ValueOrError: testValue, IsError: true}
		}
	})
}

func testPrintValue(f Future, ch chan int) {
	x, err := f.GetResult()
	if err == nil {
		fmt.Printf("TestPrintValue: %d \n", x)
		ch <- x.(int)
	}else {
		ch <- 0
		err.Error()
	}
}

func testCalculationValue(p ExplicitPromise) {
	x := 0
	for i := 0; i < 10; i++ {
		x = x + i
	}
	p.PromiseValue(x)
}

func getRequest(website string) ImplicitPromise {
	return MakeImplicitPromise(func() FutureResult {
		resp, err := http.Get(website)
		if err == nil {
			return FutureResult{ValueOrError: resp, IsError: false}
		} else{
			return FutureResult{ValueOrError: err, IsError: true}
		}

	})
}

func readIO(resp *http.Response) ImplicitPromise {
	return MakeImplicitPromise(func() FutureResult {
		bytes, err := io.ReadAll(resp.Body)
		if err == nil {
			resp.Body.Close()
			return FutureResult{bytes, false}
		} else{
			return FutureResult{err, true}
		}

	})
}