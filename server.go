package main

import (
	"apache.com/usergrid/portal"
	"apache.com/usergrid/stack"
	"apigee.com/portal"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var (
	portal       Portal
	openBrowser  bool
	startStack   bool
	stackPort    int = 8888
	re           *regexp.Regexp
	m0           *runtime.MemStats = &runtime.MemStats{}
	m1           *runtime.MemStats = &runtime.MemStats{}
	t0           time.Time
	lastModified string
	expires      string
	cmd          *exec.Cmd
)

type Portal struct {
	Host    string
	Port    int
	Backend string
	Apigee  bool
}

func GetAsset(p string) ([]byte, error) {
	if portal.Apigee {
		return apigee.Asset(p)
	} else {
		return usergrid.Asset(p)
	}
}
func (portal Portal) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p := req.URL.Path
	d := filepath.Dir(p)
	f := filepath.Base(p)
	if f == "." || f == "/" {
		f = "index.html"
	}
	d = strings.TrimPrefix(d, "/")
	if d == "" || d == "/" {
		p = filepath.Clean(fmt.Sprintf("%s", f))
	} else {
		p = filepath.Clean(fmt.Sprintf("%s/%s", d, f))
	}
	data, err := GetAsset(p)
	if err != nil || len(data) == 0 {
		log.Printf("%d\t%s", http.StatusNotFound, p)
		http.Error(w, "Not Found", http.StatusNotFound)
	} else {
		if p == "config.js" {
			log.Printf("replacing backend url with '%s'", portal.Backend)
			if re.Match(data) {
				data = re.ReplaceAll(data, []byte(fmt.Sprintf("Usergrid.overrideUrl = '%s';", portal.Backend)))
			} else {
				log.Printf("No Usergrid.overrideUrl found")
			}
		}
		mimetype := http.DetectContentType(data) //mime.TypeByExtension(filepath.Ext(p))
		if strings.HasPrefix(mimetype, "text/plain") || mimetype == "" {
			mimetype = mime.TypeByExtension(filepath.Ext(p))
		}
		if strings.HasPrefix(mimetype, "text/plain") || mimetype == "" {
			mimetype = "text/html;"
		}
		w.Header().Add("Content-Type", mimetype)
		w.Header().Add("Last-Modified", lastModified)
		w.Header().Add("Expires", expires)
		log.Printf("%d\t%s\t%s\t%d bytes", 200, mimetype, p, len(data))
		w.Write(data)
	}
}
func launch(url string) {
	time.Sleep(100 * time.Millisecond)
	cmd := exec.Command("open", url)
	if err := cmd.Run(); err != nil {
		log.Printf("Could not open browser: %v", err)
	}
}
func mkTempStack() {
	f, err := os.Create("./.usergrid.jar")
	defer f.Close()
	if err != nil && !os.IsExist(err) {
		log.Panicf("Could not create stack jar: %v", err)
	}
	jar, err := stack.Asset("usergrid.jar")
	if err != nil {
		log.Panicf("Couldn't open the packaged jar: %v", err)
	}
	_, err = f.Write(jar)
	if err != nil {
		log.Panicf("Could not populate stack jar: %v", err)
	}

}
func launchStack() {
	mkTempStack()

	if err := cmd.Run(); err != nil {
		log.Printf("Could not start the stack: %v", err)
		out, _ := cmd.CombinedOutput()
		log.Printf("Output from the stack: %s\n", out)

	}
}
func init() {
	flag.IntVar(&portal.Port, "port", 9000, "Port to serve requests on")
	flag.StringVar(&portal.Host, "host", "127.0.0.1", "Address to bind to")
	flag.StringVar(&portal.Backend, "backend", "http://api.usergrid.com", "URL of the usergrid API")
	flag.BoolVar(&portal.Apigee, "apigee", true, "Use the full apigee portal")
	flag.BoolVar(&openBrowser, "open", true, "Open a browser window on startup")
	flag.BoolVar(&startStack, "stack", true, "Start the stack, too")
	flag.Parse()
	re = regexp.MustCompile(`Usergrid\.overrideUrl\s*=[^;]+;`)
	cmd = exec.Command("java", "-jar", "-Djava.library.path=/tmp", "-Dorg.xerial.snappy.tempdir=/tmp", "./.usergrid.jar", "-db", "-init", "-nogui", "-p", fmt.Sprintf("%d", stackPort))
	expires = time.Now().Add(24 * time.Hour).Format(time.RFC1123)
	lastModified = time.Now().Format(time.RFC1123)
	if startStack {
		portal.Backend = fmt.Sprintf("http://localhost:%d/", stackPort)
	}
}

func main() {
	if startStack {
		fmt.Println("Launching the usergrid stack.")
		go launchStack()
	}
	go trap(fmt.Sprintf("Starting server on %s:%d using API backend at %s", portal.Host, portal.Port, portal.Backend))
	if openBrowser {
		go launch(fmt.Sprintf("http://%s:%d/", portal.Host, portal.Port))
	}
	http.Handle("/", portal)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", portal.Host, portal.Port), nil)
	log.Panicf(err.Error())
}

func trap(msg string) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	signal.Notify(sig, syscall.SIGTERM)
	signal.Notify(sig, syscall.SIGABRT)
	stat(msg)
	for {
		s := <-sig
		log.Printf("Received signal: %v\n", s)
		if startStack {
			log.Println("Shutting down the stack.")
			cmd.Process.Kill()
			log.Printf("Removing temp files\n")
			os.Remove("./.usergrid.jar")
		}
		stat("Exiting...")
		os.Exit(0)
	}
}
func stat(message string) {
	log.Printf(message)
	runtime.ReadMemStats(m1)
	num := runtime.NumGoroutine()
	log.Printf("Number of goroutines: %d\n", num)
	log.Printf("Current total memory: %.2fmb\n", float64(m1.Sys)/(1024*1024))
	log.Printf("Per goroutine:\n")
	log.Printf("\tMemory: %.2f bytes\n", float64(m1.Sys-m0.Sys)/float64(num))
	log.Printf("\tTime:   %.2f µs\n", float64(time.Since(t0).Nanoseconds())/float64(num)/1e3)
}
