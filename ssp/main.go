package main

import (
    "net/http"
    "time"
    "encoding/json"
    "io/ioutil"
    "bytes"
)

var DSPS = []string {
    "https://domain1.com",
    "https://domain2.com",
    "https://domain3.com",
}

const TMAX = 200 // 200 miniseconds

type Bid struct {
    BidPrice int `json:"bidprice"`
    Body string `json:"body"`
}

func GetBid(url string, w string, h string, c chan Bid) {
    values := map[string]string{"w": w, "h": h}
    jsonValue, err := json.Marshal(values)
    if err != nil {
        c <- Bid{0, ""}
        return
    }

    res, err := http.Post(url + "/bid", "application/json", bytes.NewBuffer(jsonValue))
    if err != nil {
        c <- Bid{0, ""}
        return
    }

    defer res.Body.Close()
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        c <- Bid{0, ""}
        return
    }

    bid := Bid{}
    err = json.Unmarshal(body, &bid)
    if err != nil {
        c <- Bid{0, ""}
        return
    } 

    c <- bid
}

func resp(body string, w http.ResponseWriter) {
    if body == "" {
        w.WriteHeader(http.StatusNoContent)
        w.Write([]byte("no ad"))
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte(body))
}

func Ad(w http.ResponseWriter, r *http.Request) {
    width := r.URL.Query().Get("w")
    height := r.URL.Query().Get("h")
    if width == "" || height == "" {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("params 'w' and 'h' are required"))
        return
    }

    c := make(chan Bid, 3)
    for _, dsp := range DSPS {
        go GetBid(dsp, width, height, c)
    }

    maxPrice := 0
    body := ""
    respCount := 0

    for {
        select {
        case bid := <-c:
            if bid.BidPrice > maxPrice {
                maxPrice = bid.BidPrice
                body = bid.Body
            }
            respCount += 1
            if respCount == len(DSPS) {
                resp(body, w)
                return
            }
        case <-time.After(TMAX * time.Millisecond):
            resp(body, w)
            return
        }
    }
}

func main() {
    http.HandleFunc("/ad", Ad)
    http.ListenAndServe(":8090", nil)
}
