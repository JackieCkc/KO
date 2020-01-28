package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"io/ioutil"
	"time"
	"fmt"
	"context"
	"net"
	"sort"
)

var END_POINTS = []string {
	"http://domain1.com",
	"http://domain2.com",
	"http://domain3.com",
}

func testingHTTPClient(handler http.Handler) (*http.Client, func()) {
	s := httptest.NewServer(handler)

	cli := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, network, _ string) (net.Conn, error) {
				return net.Dial(network, s.Listener.Addr().String())
			},
		},
	}

	return cli, s.Close
}

func checkStatusAndResponse(url string, expectedCode int, expectedBody string, t *testing.T) {
	SetTmax(200)
	SetDspEndpoints(END_POINTS)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetAd)
	handler.ServeHTTP(rr, req)
	if code := rr.Code; code != expectedCode {
		t.Errorf("handler returned wrong status code: got %v want %v",
			code, expectedCode)
	}
	bytes, _ := ioutil.ReadAll(rr.Body)
	if body := string(bytes); body != expectedBody {
		t.Errorf("handler returned wrong body: got '%v' want '%v'",
			body, expectedBody)
	}
}

func TestAdWithInvalidParams(t *testing.T) {
	urls := []string{"/ad", "/ad?w=1", "/ad?h=1", "/ad?w=&h=", "/ad?w=1&h=", "/ad?w=&h=1"}
	for _, url := range urls {
		checkStatusAndResponse(url, http.StatusBadRequest, "params 'w' and 'h' are required", t)
	}
}

func TestAdWithNoAd(t *testing.T) {
	checkStatusAndResponse("/ad?w=1&h=1", http.StatusNoContent, "no ad", t)
}

func TestAdWithAllDspResponed(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := "http://" + r.Host
		i := sort.StringSlice(END_POINTS).Search(domain)
		w.Write([]byte(fmt.Sprintf(`{"Bidprice": %d, "Body": "some html for domain%d"}`, i, i)))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	SetClient(httpClient)
	checkStatusAndResponse("/ad?w=1&h=1", http.StatusOK, "some html for domain2", t)
}

func TestAdWithOneDspTimeout(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := "http://" + r.Host
		if domain == END_POINTS[2] {
			time.Sleep(300 * time.Millisecond)
		}
		i := sort.StringSlice(END_POINTS).Search(domain)
		w.Write([]byte(fmt.Sprintf(`{"Bidprice": %d, "Body": "some html for domain%d"}`, i, i)))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	SetClient(httpClient)
	checkStatusAndResponse("/ad?w=1&h=1", http.StatusOK, "some html for domain1", t)
}

func TestAdWithOneDspRespondsNoContent(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := "http://" + r.Host
		if domain == END_POINTS[2] {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("not interested"))
			return
		}
		i := sort.StringSlice(END_POINTS).Search(domain)
		w.Write([]byte(fmt.Sprintf(`{"Bidprice": %d, "Body": "some html for domain%d"}`, i, i)))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	SetClient(httpClient)
	checkStatusAndResponse("/ad?w=1&h=1", http.StatusOK, "some html for domain1", t)
}

func TestAdWithAllDspTimeout(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.Write([]byte(`{"Bidprice": 1, "Body": "some html for domain"}`))
	})
	httpClient, teardown := testingHTTPClient(h)
	defer teardown()

	SetClient(httpClient)
    checkStatusAndResponse("/ad?w=1&h=1", http.StatusNoContent, "no ad", t)
}
