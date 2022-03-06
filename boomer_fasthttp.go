package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/myzhan/boomer"
	"github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var client *fasthttp.Client
var postBody []byte
var verbose bool
var method string
var url string
var timeout time.Duration
var postFile string
var rawData string
var contentType string
var jsonHeaders string
var disableKeepalive bool
var arrayHeaders []string
var csvData [][]string
var jsonValueType string
var reqCount = 1

// body = ["$a"]
// or body = {"a": "$a", "b": "$b"}
// [a$0, b$1, c$2]
// a will be replaced with the csv row, a real value will be row[0]
// b and c is the same as above
var replaceKV map[string]int
var replaceStrIndex string

func str2byte(s string) []byte {
	return []byte(s)
}

func json2map(s string) map[string]string {
	m := make(map[string]string)
	err := json.Unmarshal(str2byte(s), &m)
	if s == "" {
		return m
	}
	if err != nil {
		log.Println("parse json headers err: " + err.Error())
	}
	return m
}

// read csv to array

func csv2array(file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Println("open csv file err: " + err.Error())
		return
	}
	reader := csv.NewReader(f)
	recordCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			log.Println(fmt.Sprintf("read csv file end; record count: %d", recordCount))
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		csvData = append(csvData, record)
		recordCount++
	}
}

func string2int(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Println("parse int err: " + err.Error())
		return -1, err
	}
	return i, nil
}

// string2map convert string to map
func string2map(s string) {
	err := json.Unmarshal([]byte(s), &replaceKV)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
}

func getRow() []string {
	mod := (reqCount - 1) % len(csvData)
	row := csvData[mod]
	return row
}

type TjsonValues interface {
	int | string | []int | interface{}
}

func HandleReplaceBody(rawData []byte, dataType string, row []string, replaceKV map[string]int) []byte {
	m := make(map[string]TjsonValues)
	err := json.Unmarshal(rawData, &m)
	if err != nil {
		log.Fatalf("parse json err: " + err.Error())
	}
	var newValue TjsonValues
	for k, v := range m {
		i, ok := replaceKV[fmt.Sprintf("%s", v)]
		if !ok {
			continue
		}
		originValue := row[i]
		switch dataType {
		case "int":
			intValue, err := string2int(originValue)
			if err != nil {
				log.Println("parse int err: " + err.Error())
				continue
			}
			newValue = intValue
		case "intArray":
			intValue, err := string2int(originValue)
			if err != nil {
				log.Println("parse int err: " + err.Error())
				continue
			}
			newValue = [1]int{intValue}
		case "interface":
			intValue, err := string2int(originValue)
			if err != nil {
				newValue = originValue
			}
			newValue = intValue
		default:
			newValue = originValue
		}
		m[k] = newValue
	}
	jsonStr, _ := json.Marshal(m)
	return jsonStr
}

func worker() {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(method)
	req.Header.SetContentType(contentType)

	setArrayHeader(req, arrayHeaders)
	setJsonHeaders(req, jsonHeaders)

	if disableKeepalive {
		req.Header.SetConnectionClose()
	}

	req.SetRequestURI(url)

	if postFile != "" {
		req.SetBody(postBody)
	}
	// TODO handle array body string, int
	if jsonValueType != "" {
		row := getRow()
		b := str2byte(rawData)
		newBody := HandleReplaceBody(b, jsonValueType, row, replaceKV)
		req.SetBody(newBody)
	} else {
		if rawData != "" {
			req.SetBody(postBody)
		}
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
	boomer.RecordSuccess("http", url, elapsed.Nanoseconds()/int64(time.Millisecond), int64(len(resp.Body())))

	if verbose {
		log.Println(string(resp.Body()))
	}

	fasthttp.ReleaseRequest(req)
	reqCount++
	fasthttp.ReleaseResponse(resp)
}

func setJsonHeaders(req *fasthttp.Request, jsonStr string) {
	for k, v := range json2map(jsonStr) {
		req.Header.Set(k, v)
	}
}

func setArrayHeader(req *fasthttp.Request, h []string) {
	for _, headers := range h {
		kv := strings.Split(headers, ":")
		k := kv[0]
		v := kv[1]
		v = strings.Trim(v, " ")
		req.Header.Set(k, v)
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	arrayHeaders = append(arrayHeaders, value)
	return nil
}

var H arrayFlags

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)

	flag.StringVar(&method, "method", "GET", "HTTP method, one of GET, POST")
	flag.StringVar(&url, "url", "", "URL")
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "HTTP request timeout")
	flag.StringVar(&postFile, "post-file", "", "File containing data to POST. Remember also to set --content-type")
	flag.StringVar(&rawData, "raw-data", "", "raw data to POST. Remember also to set --content-type")

	flag.StringVar(&contentType, "content-type", "text/plain", "Content-type header")

	flag.StringVar(&jsonHeaders, "json-headers", "", "json header")
	flag.Var(&H, "H", "header arrays.")

	flag.StringVar(&jsonValueType, "json-value-type", "", `one of int, intArray, string, interface, default is ""`)
	flag.StringVar(&replaceStrIndex, "replace-str-index", "", `replace string index: '{"$a": 0}', body {"a": "$a"}, value $a replace csv per row[0]`)

	flag.BoolVar(&disableKeepalive, "disable-keepalive", false, "Disable keepalive")

	flag.BoolVar(&verbose, "verbose", false, "Print debug log")

	flag.Parse()

	log.Printf(`Fasthttp benchmark is running with these args:
method: %s
url: %s
timeout: %v
post-file: %s
raw-data: %s
replace-str-index: %s
json-value-type: %s
content-type: %s
disable-keepalive: %t
verbose: %t`, method, url, timeout, postFile, rawData, replaceStrIndex, jsonValueType, contentType, disableKeepalive, verbose)

	if url == "" {
		log.Fatalln("--url can't be empty string, please specify a URL that you want to test.")
	}

	if method != "GET" && method != "POST" {
		log.Fatalln("HTTP method must be one of GET, POST.")
	}

	if replaceStrIndex != "" {
		csv2array("./data.csv")
		// init replaceKV
		string2map(replaceStrIndex)
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
