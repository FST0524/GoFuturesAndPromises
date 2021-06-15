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
		fmt.Print(err.Error())
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
		t.Error(err.Error())
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
		t.Error(err.Error())
	}
}
//Testing Implicit Version
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

//Combined Web and IO test
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
	implicitPromise := getRequest("https://jsonlint.com/")
	future := implicitPromise.GetFuture()
	result, err := future.GetResultWithTimeout(5)
	if err == nil{
		request := result.(*http.Response)
		fmt.Printf("Expected: Request Header: %s \n",request.Header)
	}else{
		t.Errorf("Unexpected: %s \n",err.Error())
	}
}
// Test the GetResultWithTimeout function
// The timeout will happen
func TestGetResultWithOccurringTimeout(t *testing.T) {
	implicitPromise := getRequest("https://jsonlint.com/")
	future := implicitPromise.GetFuture()
	_, err := future.GetResultWithTimeout(0)
	if err == nil{
		t.Error("Unexpected: Failed there should have been a timeout error \n")
	}else{
		fmt.Printf("Expected: %s",err.Error())
	}
}

// Test the OnResolvedWithTimeout function
// SuccessFunc should be executed
func TestOnResolvedWithTimeoutSuccess(t *testing.T) {
	implicitPromise := getRequest("https://jsonlint.com/")
	future := implicitPromise.GetFuture()
	future.OnResolvedWithTimeout(func() {
		t.Errorf("Unexpected: Error occurred!")
	},
	func(i interface{}) {
		fmt.Print("Value received: \n", i)
	},
	5)
}

// Test the OnResolvedWithTimeout function
// ErrorFunc should be executed
func TestOnResolvedWithTimeoutError(t *testing.T) {
	implicitPromise := getRequest("https://jsonlint.com/")
	future := implicitPromise.GetFuture()
	future.OnResolvedWithTimeout(func() {
		fmt.Print("Expected: ErrorFunc is executing")
	},
	func(i interface{}) {
		fmt.Printf("Unexpected: No timeout occurred and value %d received.\n", i)
	},
	0)
}

// --Experimental-- Test
// This test is trying to prove whether there is a memory leak due to the for loops which may be preventing garbage collection
// MakeImplicitPromise and PromiseValue are effected
// Go Profiler will be used to analyze the result
func TestMemoryLeak(t *testing.T) {
	var y [1000]ImplicitPromise
	for i := 0; i < 1000; i++ {
		y[i] = simpleTestFunction(i)
		time.Sleep(10*time.Millisecond)
	}
	fmt.Println(y[999].GetFuture().GetResult())
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
			return FutureResult{ValueOrError: bytes, IsError: false}
		} else{
			return FutureResult{ValueOrError: err, IsError: true}
		}

	})
}

func simpleTestFunction(value int)  ImplicitPromise {
	return MakeImplicitPromise(func() FutureResult {
		if value < 10 {
			return FutureResult{ValueOrError: value, IsError: false}
		} else {
			return FutureResult{ValueOrError: PromiseError{Reason: "Input for simpleFunction has to be smaller than 10.\n"}, IsError: true}
		}
	})
}