package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/myzhan/boomer"
	"github.com/valyala/fasthttp"
)

var client *fasthttp.Client
var postBody []byte
var verbose bool
var method string
var url string
var timeout time.Duration
var postFile string
var rawData string
var contentType string
var disableKeepalive bool

func str2byte(s string) []byte {
	return []byte(s)
}

func worker() {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(method)
	req.Header.SetContentType(contentType)
	if disableKeepalive {
		req.Header.SetConnectionClose()
	}
	req.SetRequestURI(url)
	if rawData != "" {
		req.SetBody(postBody)
	}
	if postFile != "" {
		req.SetBody(postBody)
	}
	resp := fasthttp.AcquireResponse()

	startTime := time.Now()
	err := client.DoTimeout(req, resp, timeout)
	elapsed := time.Since(startTime)

	if err != nil {
		switch err {
		case fasthttp.ErrTimeout:
			boomer.RecordFailure("http", "timeout", elapsed.Nanoseconds()/int64(time.Millisecond), err.Error())
		case fasthttp.ErrNoFreeConns:
			// all Client.MaxConnsPerHost connections to the requested host are busy
			// try to increase MaxConnsPerHost
			boomer.RecordFailure("http", "connections all busy", elapsed.Nanoseconds()/int64(time.Millisecond), err.Error())
		default:
			boomer.RecordFailure("http", "unknown", elapsed.Nanoseconds()/int64(time.Millisecond), err.Error())
		}
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)
		return
	}
	fmt.Println(string(resp.Body()))
	boomer.RecordSuccess("http", url, elapsed.Nanoseconds()/int64(time.Millisecond), int64(len(resp.Body())))

	if verbose {
		log.Println(string(resp.Body()))
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.StringVar(&method, "method", "GET", "HTTP method, one of GET, POST")
	flag.StringVar(&url, "url", "", "URL")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "HTTP request timeout")
	flag.StringVar(&postFile, "post-file", "", "File containing data to POST. Remember also to set --content-type")
	flag.StringVar(&rawData, "raw-data", "", "raw data to POST. Remember also to set --content-type")
	flag.StringVar(&contentType, "content-type", "text/plain", "Content-type header")

	flag.BoolVar(&disableKeepalive, "disable-keepalive", false, "Disable keepalive")

	flag.BoolVar(&verbose, "verbose", false, "Print debug log")

	flag.Parse()

	log.Printf(`Fasthttp benchmark is running with these args:
method: %s
url: %s
timeout: %v
post-file: %s
raw-data: %s
content-type: %s
disable-keepalive: %t
verbose: %t`, method, url, timeout, postFile, rawData, contentType, disableKeepalive, verbose)

	if url == "" {
		log.Fatalln("--url can't be empty string, please specify a URL that you want to test.")
	}

	if method != "GET" && method != "POST" {
		log.Fatalln("HTTP method must be one of GET, POST.")
	}

	if method == "POST" {
		if rawData != "" {
			postBody = str2byte(rawData)
		} else {
			if postFile == "" {
				log.Fatalln("--post-file or --raw-data can't be empty string when method is POST")
			}
			tmp, err := ioutil.ReadFile(postFile)
			if err != nil {
				log.Fatalf("%v\n", err)
			}
			postBody = tmp
		}
	}

	client = &fasthttp.Client{
		MaxConnsPerHost: 2000,
	}

	task := &boomer.Task{
		Name:   "worker",
		Weight: 10,
		Fn:     worker,
	}

	boomer.Run(task)
}
