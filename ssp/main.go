package main

import (
    "net/http"
    "time"
    "encoding/json"
    "io/ioutil"
    "bytes"
    "os"
    "fmt"
)

var dspEndpoints []string
var tmax int
var client *http.Client

type Bid struct {
    BidPrice int `json:"bidprice"`
    Body string `json:"body"`
}

type Configuration struct {
    DspEndpoints []string
    Tmax int
}

func SetClient(c *http.Client) {
    client = c
}

func SetDspEndpoints(endpoints []string) {
    dspEndpoints = endpoints
}

func SetTmax(val int) {
    tmax = val
}

func loadConfig() {
    file, _ := os.Open("config.json")
    defer file.Close()
    decoder := json.NewDecoder(file)
    configuration := Configuration{}
    err := decoder.Decode(&configuration)
    if err != nil {
      fmt.Println("error: ", err)
    }
    dspEndpoints = configuration.DspEndpoints
    tmax = configuration.Tmax
}

func GetAd(w http.ResponseWriter, r *http.Request) {
    width := r.URL.Query().Get("w")
    height := r.URL.Query().Get("h")
    if width == "" || height == "" {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("params 'w' and 'h' are required"))
        return
    }

    c := make(chan Bid, len(dspEndpoints))
    for _, endpoint := range dspEndpoints {
        go getBid(endpoint, width, height, c)
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
            if respCount == len(dspEndpoints) {
                resp(body, w)
                return
            }
        case <-time.After(time.Millisecond * time.Duration(tmax)):
            resp(body, w)
            return
        }
    }
}

func getClient() *http.Client {
    if client == nil {
        client = &http.Client{}
    }

    return client
}

func getBid(url string, w string, h string, c chan Bid) {
    values := map[string]string{"w": w, "h": h}
    jsonValue, err := json.Marshal(values)
    if err != nil {
        c <- Bid{0, ""}
        return
    }

    req, err := http.NewRequest("POST", url + "/bid", bytes.NewBuffer(jsonValue))
    if err != nil {
        c <- Bid{0, ""}
        return
    }

    req.Header.Set("Content-Type", "application/json")
    res, err := getClient().Do(req)
    if err != nil || res.StatusCode != 200 {
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

func main() {
    loadConfig()
    http.HandleFunc("/ad", GetAd)
    http.ListenAndServe(":8090", nil)
}
