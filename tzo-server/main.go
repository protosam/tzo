package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"nasu/config"
	"nasu/js_engine"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// fileserver is used for routing fallback.
var fileserver = http.FileServer(http.Dir("/"))

func router(w http.ResponseWriter, r *http.Request) {
	// Measure to prevent curls from user scripts back onto server
	if !config.CONFIG.NetAllowSelf && r.Header.Get("X-Abuse-Info") == config.CONFIG.AbuseHeader {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// get config for requested hostname
	hostdir, ok := config.CONFIG.Hosts[r.Host]
	if !ok {
		// fallback to the default config or panic if we cant
		hostdir, ok = config.CONFIG.Hosts["default"]
		if !ok {
			panic("No default host configured.")
		}
	}

	// Correct hostdir to full path.
	hostdir, err := filepath.Abs(hostdir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read in host directory: " + err.Error()))
		return
	}

	// try routes.json or try if ext is tzo
	routes_path, err := filepath.Abs(hostdir + "/routes.json")
	if fileExists(routes_path) && err == nil {
		routes := make(map[string]string)

		// read in the routes.json file
		jsonFile, err := os.Open(routes_path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error reading routes.json: "))
			w.Write([]byte(err.Error()))
			return
		}
		defer jsonFile.Close()

		byteValue, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error reading routes.json: "))
			w.Write([]byte(err.Error()))
			return
		}
		json.Unmarshal(byteValue, &routes)

		// Test to find route
		for route, script := range routes {
			if args, ok := checkRoute(route, r.URL.Path); ok {
				jse := runscript(w, r, hostdir, script, args)
				time.AfterFunc(time.Duration(config.CONFIG.VMTimeoutMs)*time.Millisecond, func() {
					jse.VM.Interrupt("halt")
				})
				return
			}
		}
	} else if checkExtIsScript(r.URL.Path) { // try TZO files
		fmt.Println("RUNNING checkExtIsScript(r.URL.Path)")
		jse := runscript(w, r, hostdir, r.URL.Path, []string{})
		time.AfterFunc(time.Duration(config.CONFIG.VMTimeoutMs)*time.Millisecond, func() {
			jse.VM.Interrupt("halt")
		})
		return
	} else if r.URL.Path == "/" { // try index.tzo and main.tzo
		fmt.Println("RUNNING /")
		for _, ext := range config.CONFIG.ScriptExts {
			if len(ext) > 0 && ext[0] != '.' {
				ext = "." + ext
			}
			if fileExists(hostdir + "/index" + ext) {
				jse := runscript(w, r, hostdir, "/index"+ext, []string{})
				time.AfterFunc(time.Duration(config.CONFIG.VMTimeoutMs)*time.Millisecond, func() {
					jse.VM.Interrupt("halt")
				})
				return
			}
			if fileExists(hostdir + "/main" + ext) {
				jse := runscript(w, r, hostdir, "/main"+ext, []string{})
				time.AfterFunc(time.Duration(config.CONFIG.VMTimeoutMs)*time.Millisecond, func() {
					jse.VM.Interrupt("halt")
				})
				return
			}
		}
	}

	// fallback to static file routing if possible
	if config.CONFIG.ServeFiles && r.URL.Path != "/routes.json" && !checkExtIsScript(r.URL.Path) {
		file, err := filepath.Abs(hostdir + r.URL.Path)
		if err == nil {
			fileIsJailed, err := filepath.Match(hostdir+"/*", file)
			if fileIsJailed && err == nil && fileExists(hostdir+r.URL.Path) {
				filehandle, err := os.Open(file)
				defer filehandle.Close() //Close after function return
				if err == nil {

					//Get the Content-Type of the file
					//Create a buffer to store the header of the file in
					fileheader := make([]byte, 512)
					//Copy the headers into the FileHeader buffer
					filehandle.Read(fileheader)
					//Get content type of file
					contentType := mime.TypeByExtension(filepath.Ext(file))
					if contentType == "" { // try by sniff
						contentType = http.DetectContentType(fileheader)
					}

					//Get the file size
					fileStat, _ := filehandle.Stat()                   //Get info from file
					fileSize := strconv.FormatInt(fileStat.Size(), 10) //Get file size as a string

					//Send the headers
					//w.Header().Set("Content-Disposition", "attachment; filename="+r.URL.Query().Get("file"))
					w.Header().Set("Content-Type", contentType)
					w.Header().Set("Content-Length", fileSize)

					//Send the file
					//We read 512 bytes from the file already, so we reset the offset back to 0
					filehandle.Seek(0, 0)
					io.Copy(w, filehandle) //'Copy' the file to the client

					return
				}

			}
		}
	}

	// 404. All who end up here
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404. Page Not Found!"))
}

func runscript(w http.ResponseWriter, r *http.Request, hostdir string, script string, args []string) *js_engine.JS_ENGINE {
	w.Header().Set("Content-Type", "text/html")
	js := js_engine.New(hostdir)
	js.SetIOWriter(w)

	r.ParseForm()
	js.VM.Set("_ARGS", args)
	js.VM.Set("_HEADER", r.Header)
	js.VM.Set("_REQUEST", r.Form)
	js.VM.Set("_URL_PATH", r.URL.Path)
	js.VM.Set("_REQUEST_URI", r.RequestURI)
	js.VM.Set("_REMOTE_ADDR", r.RemoteAddr)

	js.Require(script)
	js.SendErrors()
	return js
}

func checkExtIsScript(p string) bool {
	for _, ext := range config.CONFIG.ScriptExts {
		if len(ext) > 0 && ext[0] != '.' {
			ext = "." + ext
		}
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

var conf_file = flag.String("conf", "./config.json", "What file to load config from.")

func main() {
	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for sig := range c {
			fmt.Println("Received SIGHUP, reloading config.", sig)
			config.Load("./config.json")
		}
	}()

	config.Load("./config.json")

	http.HandleFunc("/", router)

	log.Fatal(http.ListenAndServe(config.CONFIG.Bind, nil))
}

func checkRoute(route, path string) ([]string, bool) {
	args := make([]string, 0)

	if len(route) > 1 && route[len(route)-1] == '/' {
		route = route[:len(route)-1]
	}

	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	route_x := strings.Split(route, "/")
	path_x := strings.Split(path, "/")

	if len(route_x) != len(path_x) {
		return nil, false
	}
	for i := range route_x {
		if route_x[i] != path_x[i] && route_x[i] != "%" {
			return nil, false
		}
		if route_x[i] == "%" {
			args = append(args, path_x[i])
		}
	}
	return args, true
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
