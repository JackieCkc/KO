package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"gopkg.in/jarcoal/httpmock.v1"
	"io/ioutil"
	"time"
)

func TestAdWithInvalidParams(t *testing.T) {
	urls := []string{"/ad", "/ad?w=1", "/ad?h=1", "/ad?w=1&h="}
	for _, url := range urls {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(Ad)
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}
		bytes, _ := ioutil.ReadAll(rr.Body)
		if body := string(bytes); body != "params 'w' and 'h' are required" {
			t.Errorf("handler returned wrong body: got '%v' want '%v'",
				body, "params 'w' and 'h' are required")
		}
	}
}

func TestAdWithNoAd(t *testing.T) {
	req, err := http.NewRequest("GET", "/ad?w=1&h=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Ad)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
	bytes, _ := ioutil.ReadAll(rr.Body)
	if body := string(bytes); body != "no ad" {
		t.Errorf("handler returned wrong body: got '%v' want '%v'",
			body, "no ad")
	}
}

func TestAdWithAllDspResponed(t *testing.T) {
	getJsonMockBytes := func(url string) []byte {
	    switch url {
	    case "https://domain1.com/bid":
	        return []byte(`{"Bidprice": 1, "Body": "some html for domain1"}`)
	    case "https://domain2.com/bid":	
	        return []byte(`{"Bidprice": 2, "Body": "some html for domain2"}`)
	    case "https://domain3.com/bid":	
	        return []byte(`{"Bidprice": 3, "Body": "some html for domain3"}`)
	    }
	    return nil
	}

	httpmock.Activate()
    defer httpmock.DeactivateAndReset()
    for _, dsp := range DSPS {
    	url := dsp + "/bid"
        httpmock.RegisterResponder("POST", url, httpmock.NewBytesResponder(200, getJsonMockBytes(url)))
    }

	req, err := http.NewRequest("GET", "/ad?w=1&h=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Ad)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	bytes, _ := ioutil.ReadAll(rr.Body)
	if body := string(bytes); body != "some html for domain3" {
		t.Errorf("handler returned wrong body: got '%v' want '%v'",
			body, "some html for domain3")
	}
}

func TestAdWithOneDspTimeout(t *testing.T) {
	getJsonMockBytes := func(url string) []byte {
	    switch url {
	    case "https://domain1.com/bid":
	        return []byte(`{"Bidprice": 1, "Body": "some html for domain1"}`)
	    case "https://domain2.com/bid":	
	        return []byte(`{"Bidprice": 2, "Body": "some html for domain2"}`)
	    case "https://domain3.com/bid":	
	        time.Sleep(300 * time.Millisecond)
	    }
	    return nil
	}

	httpmock.Activate()
    defer httpmock.DeactivateAndReset()
    for _, dsp := range DSPS {
    	url := dsp + "/bid"
        httpmock.RegisterResponder("POST", url, httpmock.NewBytesResponder(200, getJsonMockBytes(url)))
    }

	req, err := http.NewRequest("GET", "/ad?w=1&h=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Ad)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	bytes, _ := ioutil.ReadAll(rr.Body)
	if body := string(bytes); body != "some html for domain2" {
		t.Errorf("handler returned wrong body: got '%v' want '%v'",
			body, "some html for domain2")
	}
}

func TestAdWithAllDspTimeout(t *testing.T) {
	getJsonMockBytes := func(url string) []byte {
	    switch url {
	    case "https://domain1.com/bid":
	    case "https://domain2.com/bid":	
	    case "https://domain3.com/bid":	
	    	time.Sleep(300 * time.Millisecond)
	    }
	    return nil
	}

	httpmock.Activate()
    defer httpmock.DeactivateAndReset()
    for _, dsp := range DSPS {
    	url := dsp + "/bid"
        httpmock.RegisterResponder("POST", url, httpmock.NewBytesResponder(200, getJsonMockBytes(url)))
    }

	req, err := http.NewRequest("GET", "/ad?w=1&h=1", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Ad)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
	bytes, _ := ioutil.ReadAll(rr.Body)
	if body := string(bytes); body != "no ad" {
		t.Errorf("handler returned wrong body: got '%v' want '%v'",
			body, "no ad")
	}
}