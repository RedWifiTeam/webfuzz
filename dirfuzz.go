package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	rua "github.com/EDDYCJY/fake-useragent"
)

const (
	// SimilarDistance 相似性汉明距离
	SimilarDistance = 3
	// BruteTimeout 爆破队列阻塞超时时间
	BruteTimeout = 5
)

// ReqUrls ReqUrls
type ReqUrls struct {
	t int // 1 file 2 dir
	p string
}

// Results 爆破结果结构
type Results struct {
	Forbidden []string `json:"403"`
	Ok        []string `json:"200"`
}

// Resp Resp
type Resp struct {
	Status int
	Length int
	Body   string
}

var (
	// HTTPClient HTTP 请求客户端
	HTTPClient http.Client
	// HTTPMethod HTTP 請求方法
	HTTPMethod int
	// HTTPAction HTTP 請求動作
	HTTPAction int
	// SimilarBody 相似性-404 頁面內容
	SimilarBody string

	reqUrls    chan ReqUrls
	dictUrls   chan ReqUrls
	cancelChan chan bool

	// ResultUrls 爆破结果
	ResultUrls *Results
)

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	HTTPClient = http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: tr,
	}

	HTTPMethod = 2
	HTTPAction = 1

	ResultUrls = new(Results)

	reqUrls = make(chan ReqUrls, 102400)
	dictUrls = make(chan ReqUrls, 1024)

	cancelChan = make(chan bool)
}

// HTTPRequest HTTP 请求
func HTTPRequest(url string, method int, action int) (status int, length int, body string) {
	_method := ""
	body = ""
	length = 0
	var _body io.Reader
	_body = nil
	if method == HEAD {
		_method = "HEAD"
		if action == LENGTH {
			_body = strings.NewReader("233")
		}
	} else if method == GET {
		_method = "GET"
	} else if method == POST {
		_method = "POST"
	}
	req, err := http.NewRequest(_method, url, _body)
	if err == nil {
		randomUserAgent := rua.Computer()
		req.Header.Set("User-Agent", randomUserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Connection", "close")
		if method == GET && action == RANGE {
			_r := r.Intn(9) + 1
			req.Header.Set("Range", fmt.Sprintf("bytes=-%d", _r))
		}
		resp, err := HTTPClient.Do(req)
		if err != nil {
			Error.Println("HTTPRequest Error:", err.Error())
			return 0, 0, ""
		}
		if resp.StatusCode == 302 {
			body = resp.Header.Get("Location")
			// fmt.Println(resp.Header)
			return resp.StatusCode, 0, body
		}
		defer resp.Body.Close()
		if method > HEAD && action >= NORMAL {
			_body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				Error.Println("ioutil ReadAll failed :", err.Error())
				return resp.StatusCode, length, ""
			}
			body = string(_body)
			// fmt.Println(body)
			// return resp.Status, len(body), string(body)
		}
		// Header
		if action == RANGE {
			ret := resp.Header.Get("Content-Range")
			if ret != "" {
				_tmp := strings.Split(ret, "/")
				// fmt.Println("length", _tmp)
				if len(_tmp) > 1 {
					length, _ = strconv.Atoi(_tmp[1])
				}
			}
		}
		// Content-Length
		if length == 0 {
			length, _ = strconv.Atoi(resp.Header.Get("Content-Length"))
		}
		// Calc body
		if length == 0 {
			length = len(body)
		}
		return resp.StatusCode, length, body
	}
	Error.Println("NewRequest Error:" + err.Error())
	return 400, 0, ""
}

// PrepareForBrute 预处理
func PrepareForBrute(method int, action int) bool {
	// 测试爆破方法
	/*
		Bursting Performances in Blind SQL Injection - Take 2 (Bandwidth)
		http://www.wisec.it/sectou.php?id=472f952d79293

		mHeadLength := false
		mGetRange := false
		// test for head range
		resSc, resLen, resBody = HTTPRequest(BaseURL, HEAD, LENGTH)
		if resSc == 200 && resLen > 0 {
			log.Println("[S] Good Job For HEAD & LENGTH")
			mHeadLength = true
		} else if resSc == 400 {
			log.Println("[I] Server not suppost HEAD data for 'Content-Length'")
		}
		fmt.Println("[D] HEAD LENGTH", resSc, resLen, resBody)
		// test for get range
		resSc, resLen, resBody = HTTPRequest(BaseURL, GET, RANGE)
		if resSc == 206 && resLen > 0 {
			log.Println("[S] Good Job For GET & RANGE")
			mGetRange = true
		} else {
			log.Println("[I] Server not suppost GET 'Range'")
		}
		fmt.Println("[D] GET RANGE", resSc, resLen, resBody)

		fmt.Println(mHeadLength, mGetRange)
	*/
	// test for head normal
	r := make(map[int]Resp, 5)
	// 正常页面
	Debug.Println("Try to req HEAD & NORMAL 0 : ", BaseURL)
	rS, rL, rB := HTTPRequest(BaseURL, HEAD, NORMAL)
	r[0] = Resp{rS, rL, rB}
	if rS == 200 {
		// 404
		url := fmt.Sprintf("%s/%s", BaseURL, RandString(10))
		Debug.Println("Try to req HEAD & NORMAL 1 : ", url)
		rS, rL, rB = HTTPRequest(url, HEAD, NORMAL)
		r[1] = Resp{rS, rL, rB}
		url = fmt.Sprintf("%s/%s/%s", BaseURL, RandString(10), RandString(10))
		Debug.Println("Try to req HEAD & NORMAL 2 : ", url)
		rS, rL, rB = HTTPRequest(url, HEAD, NORMAL)
		r[2] = Resp{rS, rL, rB}
		url = fmt.Sprintf("%s/%s/%s.html", BaseURL, RandString(10), RandString(5))
		Debug.Println("Try to req HEAD & NORMAL 3 : ", url)
		rS, rL, rB = HTTPRequest(url, HEAD, NORMAL)
		r[3] = Resp{rS, rL, rB}
		if (r[1].Status == r[2].Status) && (r[2].Status == r[3].Status) && (r[3].Status == 404) {
			Debug.Println("Good Job For HEAD & NORMAL")
			HTTPMethod = HEAD
			HTTPAction = NORMAL
		} else {
			Debug.Println("Try to req GET & NORMAL 0 : ", BaseURL)
			rS, rL, rB = HTTPRequest(BaseURL, GET, NORMAL)
			r[0] = Resp{rS, rL, rB}
			url = fmt.Sprintf("%s/%s", BaseURL, RandString(10))
			Debug.Println("Try to req GET & NORMAL 1 : ", url)
			rS, rL, rB = HTTPRequest(url, GET, NORMAL)
			r[1] = Resp{rS, rL, rB}
			url = fmt.Sprintf("%s/%s/%s", BaseURL, RandString(10), RandString(10))
			Debug.Println("Try to req GET & NORMAL 2 : ", url)
			rS, rL, rB = HTTPRequest(url, GET, NORMAL)
			r[2] = Resp{rS, rL, rB}
			url = fmt.Sprintf("%s/%s/%s.html", BaseURL, RandString(10), RandString(5))
			Debug.Println("Try to req GET & NORMAL 3 : ", url)
			rS, rL, rB = HTTPRequest(url, GET, NORMAL)
			r[3] = Resp{rS, rL, rB}
			if (r[2].Status == r[3].Status) && (r[3].Status == r[4].Status) && (r[4].Status == 404) {
				Debug.Println("Good Job For GET & NORMAL")
				HTTPMethod = GET
				HTTPAction = NORMAL
			} else if (r[2].Status == r[3].Status) && (r[3].Status == r[4].Status) && (r[4].Status == 200) {
				// 狀態碼為 200 的 Not Found
				// 相似性判斷
				if (r[1].Length != r[2].Length) && (r[2].Length != r[3].Length) {
					distance := [3]int{0, 0, 0}
					if math.Abs(float64(r[1].Length-r[2].Length)) < 100 {
						distance[0] = GetSimHashSimilar(r[1].Body, r[2].Body)
					}
					if math.Abs(float64(r[3].Length-r[2].Length)) < 100 {
						distance[1] = GetSimHashSimilar(r[3].Body, r[2].Body)
					}
					if math.Abs(float64(r[3].Length-r[1].Length)) < 100 {
						distance[2] = GetSimHashSimilar(r[3].Body, r[1].Body)
					}
					for _, v := range distance {
						if v > SimilarDistance {
							Debug.Println("[W] There are some warining!")
							Debug.Println("1x2", distance[0], "3x2", distance[1], "3x1", distance[2])
							return false
						}
					}
				}

			}
		}
	}
	return true
}

// StartToBrute 開始爆破 - 動態線程處理
func StartToBrute(wg *sync.WaitGroup, webType string, threadCount int, interval int, timeout time.Duration) {
	Debug.Println("StartToBrute...")
	defer wg.Done()
	//并发访问网址
	for i := 0; i < threadCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-time.After(timeout):
					if cancelChan != nil {
						cancelChan <- true
					}
					return
				case <-cancelChan:
					return
				case uri := <-reqUrls:
					_uri := StringStrip(uri.p, "/")
					if uri.t == 2 && !strings.HasSuffix(_uri, "/") {
						_uri += "/"
					}
					url := strings.Join([]string{BaseURL, _uri}, "/")
					// Debug.Println(uri, url)
					rS, _, _ := HTTPRequest(url, HTTPMethod, HTTPAction)
					// fmt.Printf("[*] %d - %-100s\n", rS, url)
					fmt.Printf("\r[*] %d - %-100s\r", rS, url)
					if rS == 200 {
						ResultUrls.Ok = append(ResultUrls.Ok, url)
						if uri.t == 2 {
							_p := ReqUrls{2, _uri}
							dictUrls <- _p
						}
						// fmt.Printf("[*] %d - %-100s\n", rS, url)
						fmt.Printf("\r[*] %d - %-100s\n", rS, url)
					} else if rS == 403 {
						if uri.t == 2 && webType == "jsp" {
							_p := ReqUrls{2, _uri}
							dictUrls <- _p
						}
						ResultUrls.Forbidden = append(ResultUrls.Forbidden, url)
						// fmt.Printf("[*] %d - %-100s\n", rS, url)
						fmt.Printf("\r[*] %d - %-100s\n", rS, url)
					}
					// 间隔时间
					if interval > 0 {
						time.Sleep(time.Duration(interval) * time.Second)
					}
					// time.Sleep(time.Duration(1) * time.Second)
				}
			}
		}()
	}
}

// GetDictUrls 从自动获取并处理地址
func GetDictUrls(wg *sync.WaitGroup) {
	Debug.Println("GetDictUrls ...")
	defer wg.Done()
	for {
		select {
		case <-cancelChan:
			return
		case uri := <-dictUrls:
			_uri := StringStrip(uri.p, "/")
			paths := strings.Split(_uri, "/")
			Debug.Println("GetDictUrls Uri.p : ", _uri)
			Debug.Println("GetDictUrls paths : ", paths)
			t := Nodes
			// TODO ...
			for i, p := range paths {
				Debug.Println("GetDictUrls range paths p : ", p)
				if _, ok := t.Nodes[p]; ok {
					t = t.getNode(p)
				}
				// 最后一个路径
				if i == len(paths)-1 && t.Path == p {
					for _, u := range t.getFiles() {
						_q := ReqUrls{1, fmt.Sprintf("%s/%s", _uri, u)}
						reqUrls <- _q
					}
					for _, u := range t.getNodeKeys() {
						_q := ReqUrls{2, fmt.Sprintf("%s/%s/", _uri, u)}
						reqUrls <- _q
					}
				}
			}
		}
	}
}

func saveResult(name string) {
	data, err := json.MarshalIndent(&ResultUrls, "", "\t")
	if err != nil {
		Error.Println("Marshal ResultUrls : ", err)
	}
	if ioutil.WriteFile(name, data, 0644) == nil {
		// log.Println(string(data))
		Info.Println("[+] Save to results.log")
	}
}

// Dispatcher Fuzz 调度器
func Dispatcher(threadCount, interval int, webType, dbFile string) {
	Debug.Println("Dispatcher...")

	wg := &sync.WaitGroup{}

	dictNode := Nodes
	dictNode.load(dbFile)

	// 插入初始地址
	Debug.Println("Dispatcher Insert reqUrls...")
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, d := range dictNode.getNodeKeys() {
			_q := ReqUrls{2, StringStrip(d, "/")}
			reqUrls <- _q
		}
		for _, d := range dictNode.getFiles() {
			_q := ReqUrls{1, StringStrip(d, "/")}
			reqUrls <- _q
		}
		Info.Println("[+] Load init dict : ", len(reqUrls))
	}()
	Debug.Println("Dispatcher GetDictUrls...")

	wg.Add(1)
	go GetDictUrls(wg)
	Debug.Println("Dispatcher StartToBrute...")

	wg.Add(1)
	go StartToBrute(wg, webType, threadCount, interval, 5*time.Second)

	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case v := <-cancelChan:
			if v {
				close(cancelChan)
			}
		}
	}()

	wg.Wait()
	Debug.Println("Dispatcher Finish...")

	// ResultUrls
	sort.Strings(ResultUrls.Ok)
	sort.Strings(ResultUrls.Forbidden)
	Info.Printf("[+] Result :\t[200] : %d\t[403] : %d\n", len(ResultUrls.Ok), len(ResultUrls.Forbidden))
	saveResult("result.log")
}
