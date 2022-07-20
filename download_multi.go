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

type Downloader struct {
	concurrency   int              //No. of connections
	uri           string           //URL of the file we want to download
	chunks        map[int]*os.File //Map of temporary files we are creating
	err           error            //used when error occurs inside a goroutine
	startTime     time.Time        //to track time took
	fileName      string           //name of the file we are downloading
	out           *os.File         //output / downloaded file
	to            string
	progressBar   map[int]*progress //index => progress
	stop          chan error        //to handle stop signals from terminal
	*sync.RWMutex                   //mutex to lock the maps which accessing it concurrently
	progress      func(now, total int, percent float64)
	header        map[string]string
	breaks        chan bool
	totalTime     time.Duration
}

type progress struct {
	curr  int //curr is the current read till now
	total int //total bytes which we are supposed to read
}

func DownloadFast(from, to string) (a *Downloader) {
	a = new(Downloader)
	a.to = to
	a.breaks = make(chan bool)
	a.concurrency = runtime.NumCPU()
	a.uri = from
	a.chunks = make(map[int]*os.File)
	a.startTime = time.Now()
	a.fileName = filepath.Base(a.uri)
	a.RWMutex = &sync.RWMutex{}
	a.progressBar = make(map[int]*progress)
	a.stop = make(chan error)
	a.header = make(map[string]string)
	return
}

func (a *Downloader) Header(k, v string) *Downloader {
	a.header[k] = v
	return a
}

func (a *Downloader) Stop() {
	a.breaks <- true
}

func (a *Downloader) Start(progress ...func(now, total int, percent float64)) (err error) {

	if progress != nil {
		a.progress = progress[0]
	}

	if err = a.createOutputFile(a.to); err != nil {
		return
	}

	go a.catchSignals()
	go a.userstop(a.breaks)

	if err = a.run(); err != nil {
		return
	}

	a.totalTime = time.Since(a.startTime)
	return

}

func (a *Downloader) userstop(s chan bool) {
	<-s
	for i := 0; i < len(a.chunks); i++ {
		a.stop <- fmt.Errorf("stop")
	}
}

//createOutputFile ...
func (a *Downloader) createOutputFile(to string) (err error) {
	out, err := os.OpenFile(to, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return
	}
	a.out = out
	return nil
}

//run is basically the start method
func (a *Downloader) run() error {

	support, contentLength, err := a.getRangeDetails(a.uri)
	if err != nil {
		return err
	}

	if !support {
		a.concurrency = 1
	}

	return a.process(contentLength)

}

//process is the manager method
func (a *Downloader) process(contentLength int) error {

	//Close the output file after everything is done
	defer a.out.Close()

	split := contentLength / a.concurrency

	wg := &sync.WaitGroup{}
	index := 0

	for i := 0; i < contentLength; i += split + 1 {
		j := i + split
		if j > contentLength {
			j = contentLength
		}

		f, err := os.CreateTemp("", a.fileName+".*.part")
		if err != nil {
			return err
		}
		defer f.Close()
		defer os.Remove(f.Name())

		a.chunks[index] = f
		a.progressBar[index] = &progress{curr: 0, total: j - i}

		wg.Add(1)
		go a.downloadFileForRange(wg, a.uri, strconv.Itoa(i)+"-"+strconv.Itoa(j), index, f)
		index++
	}

	stop := make(chan struct{})

	//Keep Printing Progress
	go a.startProgressBar(stop)
	wg.Wait()

	stop <- struct{}{}

	if a.err != nil {
		os.Remove(a.out.Name())
		return a.err
	}

	return a.combineChunks()
}

func (a *Downloader) startProgressBar(stop chan struct{}) {

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var count, total int
			for i := 0; i < len(a.progressBar); i++ {
				a.RLock()
				p := *a.progressBar[i]
				count = count + p.curr
				total = total + p.total
				a.RUnlock()
			}
			a.progress(count, total, math.Floor(float64(count)/float64(total)*100))
		case <-stop:
			var count, total int
			for i := 0; i < len(a.progressBar); i++ {
				a.RLock()
				p := *a.progressBar[i]
				count = count + p.curr
				total = total + p.total
				a.RUnlock()
				a.progress(count, total, math.Floor(float64(count)/float64(total)*100))
			}
			return
		}
	}

}

//combineChunks will combine the chunks in ordered fashion starting from 1
func (a *Downloader) combineChunks() error {

	var w int64
	//maps are not ordered hence using for loop
	for i := 0; i < len(a.chunks); i++ {
		handle := a.chunks[i]
		handle.Seek(0, 0) //We need to seek because read and write cursor are same and the cursor would be at the end.
		written, err := io.Copy(a.out, handle)
		if err != nil {
			return err
		}
		w += written
	}

	//log.Printf("Wrote to File : %v, Written bytes : %v", a.out.Name(), w)

	return nil
}

//downloadFileForRange will download the file for the provided range and set the bytes to the chunk map, will set summor.error field if error occurs
func (a *Downloader) downloadFileForRange(wg *sync.WaitGroup, u, r string, index int, handle io.Writer) {

	defer wg.Done()

	request, err := http.NewRequest("GET", u, strings.NewReader(""))
	if err != nil {
		a.err = err
		return
	}

	request.Header.Add("Range", "bytes="+r)

	for k, v := range a.header {
		request.Header.Add(k, v)
	}

	sc, err := a.getDataAndWriteToFile(request, handle, index)
	if err != nil {
		a.err = err
		return
	}

	//206 = Partial Content
	if sc != 200 && sc != 206 {
		a.Lock()
		a.err = fmt.Errorf("Did not get 20X status code, got : %v", sc)
		a.Unlock()
		log.Println(a.err)
		return
	}

}

//getRangeDetails returns ifRangeIsSupported,statuscode,error
func (a *Downloader) getRangeDetails(u string) (bool, int, error) {

	request, err := http.NewRequest("HEAD", u, strings.NewReader(""))
	if err != nil {
		return false, 0, fmt.Errorf("Error while creating request : %v", err)
	}

	sc, headers, _, err := a.doAPICall(request)
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
func (a *Downloader) doAPICall(request *http.Request) (int, http.Header, []byte, error) {

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
func (a *Downloader) getDataAndWriteToFile(request *http.Request, f io.Writer, index int) (int, error) {

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
		case cErr := <-a.stop:
			return response.StatusCode, cErr
		default:
			err := a.readBody(response, f, buf, &readTotal, index)
			if err == io.EOF {
				return response.StatusCode, nil
			}

			if err != nil {
				return response.StatusCode, err
			}
		}
	}
}

func (a *Downloader) readBody(response *http.Response, f io.Writer, buf []byte, readTotal *int, index int) error {

	r, err := response.Body.Read(buf)

	if r > 0 {
		f.Write(buf[:r])
	}

	if err != nil {
		return err
	}

	*readTotal += r

	a.Lock()
	a.progressBar[index].curr = *readTotal
	a.Unlock()

	return nil
}

func (a *Downloader) catchSignals() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		s := <-sigc
		for i := 0; i < len(a.chunks); i++ {
			a.stop <- fmt.Errorf("got stop signal : %v", s)
		}
	}()
}
