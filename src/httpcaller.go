package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	URL         = "http://localhost:8880/ping"
	contentType = "application/x-www-form-urlencoded"
)

var (
	reqPerM  uint64
	routines = 1
)

const (
	TIMEOUT = time.Second * 5
)

type Result struct {
	//failed sending request
	Success bool
	//response status
	StatusCode    int
	LatencyMillis float64
}

func init() {
	flag.Uint64Var(&reqPerM, "ReqPMin", 600, "Max amount of requests")
	flag.Parse()
}

func CreateSessionId() (string, bool) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", false
	}
	return base64.URLEncoding.EncodeToString(b), true
}

func HttpPostRequest() {
	result := make(chan Result)
	go func() {
		select {
		case <-time.After(TIMEOUT):
			log.Println("HttpPostRequest: Timeout occurred, server could potentially lose message")
		case r := <-result:
			log.Println("HttpPostRequest: Server send response", r)
		}
	}()
	start := time.Now()
	data := url.Values{}
	data.Set("domain", "mitom.tv")

	session, _ := CreateSessionId()
	data.Set("session", session)

	location, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	timeStamp := time.Now().In(location).Format("2006-01-02 15:04:05")
	data.Set("updateat", timeStamp)

	resp, err := http.Post(URL, contentType, strings.NewReader(data.Encode()))
	if err != nil {
		//under heavy load it might happen
		//tcp: lookup localhost: device or resource busy
		log.Println("HttpPostRequest: Couldn't send a request to server.", err)
		result <- Result{false, 0, 0}
		return
	}
	result <- Result{true, resp.StatusCode, float64(time.Since(start).Seconds() * 1000)}
	defer resp.Body.Close()
}

// ExecuteRequests test
func ExecuteRequests() {
	numProcs := routines * runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(numProcs)
	for p := 0; p < numProcs; p++ {
		go func() {
			defer wg.Done()
			//uniformly send requests
			limiter := time.Tick(time.Minute / time.Duration(reqPerM/uint64(numProcs)))
			for int64(atomic.AddUint64(&reqPerM, ^uint64(0))) >= 0 {
				<-limiter
				go HttpPostRequest()
			}
		}()
	}
	wg.Wait()
}

func main_http() {
	ExecuteRequests()
}
