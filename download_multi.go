package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type downloader struct {
	concurrency   int               //No. of connections
	uri           string            //URL of the file we want to download
	chunks        map[int]*os.File  //Map of temporary files we are creating
	err           error             //used when error occurs inside a goroutine
	startTime     time.Time         //to track time took
	fileName      string            //name of the file we are downloading
	out           *os.File          //output / downloaded file
	progressBar   map[int]*progress //index => progress
	stop          chan error        //to handle stop signals from terminal
	*sync.RWMutex                   //mutex to lock the maps which accessing it concurrently
	progress      func(now, total int, percent float64)
}

type progress struct {
	curr  int //curr is the current read till now
	total int //total bytes which we are supposed to read
}

// по дефолту max cores
/*
	Downloader(link, "out.zip", func(now, total int, percent float64) {
		console.Play("%dMb/%dMb — %%%v", now/1024/1024, total/1024/1024, percent)
	})
*/
func Downloader(from, to string, progressHandler func(now, total int, percent float64), stops ...chan bool) (totalTime time.Duration, err error) {

	sum := new(downloader)
	sum.concurrency = runtime.NumCPU()
	sum.progress = progressHandler
	sum.uri = from
	sum.chunks = make(map[int]*os.File)
	sum.startTime = time.Now()
	sum.fileName = filepath.Base(sum.uri)
	sum.RWMutex = &sync.RWMutex{}
	sum.progressBar = make(map[int]*progress)
	sum.stop = make(chan error)

	if err = sum.createOutputFile(to); err != nil {
		return
	}

	//get the user kill signals
	go sum.catchSignals()
	if stops != nil {
		go sum.userstop(stops[0])
	}

	if err = sum.run(); err != nil {
		return
	}

	totalTime = time.Since(sum.startTime)
	return

}

func (sum *downloader) userstop(s chan bool) {
	<-s
	for i := 0; i < len(sum.chunks); i++ {
		sum.stop <- fmt.Errorf("stop")
	}
}

//createOutputFile ...
func (sum *downloader) createOutputFile(to string) (err error) {
	out, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return
	}
	sum.out = out
	return nil
}

//run is basically the start method
func (sum *downloader) run() error {

	support, contentLength, err := getRangeDetails(sum.uri)
	if err != nil {
		return err
	}

	if !support {
		sum.concurrency = 1
	}

	return sum.process(contentLength)

}

//process is the manager method
func (sum *downloader) process(contentLength int) error {

	//Close the output file after everything is done
	defer sum.out.Close()

	split := contentLength / sum.concurrency

	wg := &sync.WaitGroup{}
	index := 0

	for i := 0; i < contentLength; i += split + 1 {
		j := i + split
		if j > contentLength {
			j = contentLength
		}

		f, err := os.CreateTemp("", sum.fileName+".*.part")
		if err != nil {
			return err
		}
		defer f.Close()
		defer os.Remove(f.Name())

		sum.chunks[index] = f
		sum.progressBar[index] = &progress{curr: 0, total: j - i}

		wg.Add(1)
		go sum.downloadFileForRange(wg, sum.uri, strconv.Itoa(i)+"-"+strconv.Itoa(j), index, f)
		index++
	}

	stop := make(chan struct{})

	//Keep Printing Progress
	go sum.startProgressBar(stop)
	wg.Wait()

	stop <- struct{}{}

	if sum.err != nil {
		os.Remove(sum.out.Name())
		return sum.err
	}

	return sum.combineChunks()
}

func (sum *downloader) startProgressBar(stop chan struct{}) {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var count, total int
			for i := 0; i < len(sum.progressBar); i++ {
				sum.RLock()
				p := *sum.progressBar[i]
				count = count + p.curr
				total = total + p.total
				sum.RUnlock()
			}
			sum.progress(count, total, math.Floor(float64(count)/float64(total)*100))
		case <-stop:
			var count, total int
			for i := 0; i < len(sum.progressBar); i++ {
				sum.RLock()
				p := *sum.progressBar[i]
				count = count + p.curr
				total = total + p.total
				sum.RUnlock()
				sum.progress(count, total, math.Floor(float64(count)/float64(total)*100))
			}
			return
		}
	}

}

//combineChunks will combine the chunks in ordered fashion starting from 1
func (sum *downloader) combineChunks() error {

	var w int64
	//maps are not ordered hence using for loop
	for i := 0; i < len(sum.chunks); i++ {
		handle := sum.chunks[i]
		handle.Seek(0, 0) //We need to seek because read and write cursor are same and the cursor would be at the end.
		written, err := io.Copy(sum.out, handle)
		if err != nil {
			return err
		}
		w += written
	}

	//log.Printf("Wrote to File : %v, Written bytes : %v", sum.out.Name(), w)

	return nil
}

//downloadFileForRange will download the file for the provided range and set the bytes to the chunk map, will set summor.error field if error occurs
func (sum *downloader) downloadFileForRange(wg *sync.WaitGroup, u, r string, index int, handle io.Writer) {

	defer wg.Done()

	request, err := http.NewRequest("GET", u, strings.NewReader(""))
	if err != nil {
		sum.err = err
		return
	}

	request.Header.Add("Range", "bytes="+r)

	sc, err := sum.getDataAndWriteToFile(request, handle, index)
	if err != nil {
		sum.err = err
		return
	}

	//206 = Partial Content
	if sc != 200 && sc != 206 {
		sum.Lock()
		sum.err = fmt.Errorf("Did not get 20X status code, got : %v", sc)
		sum.Unlock()
		log.Println(sum.err)
		return
	}

}

//getRangeDetails returns ifRangeIsSupported,statuscode,error
func getRangeDetails(u string) (bool, int, error) {

	request, err := http.NewRequest("HEAD", u, strings.NewReader(""))
	if err != nil {
		return false, 0, fmt.Errorf("Error while creating request : %v", err)
	}

	sc, headers, _, err := doAPICall(request)
	if err != nil {
		return false, 0, fmt.Errorf("Error calling url : %v", err)
	}

	if sc != 200 && sc != 206 {
		return false, 0, fmt.Errorf("Did not get 200 or 206 response")
	}

	conLen := headers.Get("Content-Length")
	cl, err := strconv.Atoi(conLen)
	if err != nil {
		return false, 0, fmt.Errorf("Error Parsing content length : %v", err)
	}

	//Accept-Ranges: bytes
	if headers.Get("Accept-Ranges") == "bytes" {
		return true, cl, nil
	}

	return false, cl, nil

}

//doAPICall will do the api call and return statuscode,headers,data,error respectively
func doAPICall(request *http.Request) (int, http.Header, []byte, error) {

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, http.Header{}, []byte{}, fmt.Errorf("Error while doing request : %v", err)
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, http.Header{}, []byte{}, fmt.Errorf("Error while reading response body : %v", err)
	}

	return response.StatusCode, response.Header, data, nil

}

//getDataAndWriteToFile will get the response and write to file
func (sum *downloader) getDataAndWriteToFile(request *http.Request, f io.Writer, index int) (int, error) {

	client := http.Client{
		Timeout: 0,
	}

	response, err := client.Do(request)
	if err != nil {
		return response.StatusCode, fmt.Errorf("Error while doing request : %v", err)
	}
	defer response.Body.Close()

	//we make buffer of 500 bytes and try to read 500 bytes every iteration.
	var buf = make([]byte, 500)
	var readTotal int

	for {
		select {
		case cErr := <-sum.stop:
			return response.StatusCode, cErr
		default:
			err := sum.readBody(response, f, buf, &readTotal, index)
			if err == io.EOF {
				return response.StatusCode, nil
			}

			if err != nil {
				return response.StatusCode, err
			}
		}
	}
}

func (sum *downloader) readBody(response *http.Response, f io.Writer, buf []byte, readTotal *int, index int) error {

	r, err := response.Body.Read(buf)

	if r > 0 {
		f.Write(buf[:r])
	}

	if err != nil {
		return err
	}

	*readTotal += r

	sum.Lock()
	sum.progressBar[index].curr = *readTotal
	sum.Unlock()

	return nil
}

func (sum *downloader) catchSignals() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		for i := 0; i < len(sum.chunks); i++ {
			sum.stop <- fmt.Errorf("got stop signal : %v", s)
		}
	}()
}
