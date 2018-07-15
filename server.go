//
// server.go is a simple HTTP server which lets you view and launch jobs.
//
// When you access http://localhost:8080/ you will see a list of all the
// recipes available - which is nothing more than a glob of <*.recipe>.
//
// If you click upon one of the jobs it will be immediately executed,
// and you will see the output streamed to your browser.  As more output
// appears you will be scrolled down to follow it automatically.
//
// Once the build has completed you will be alert()'d.
//
// This is another quick hack..
//
// Steve
// --

package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// IndexHandler shows all available recipes/jobs, and serves a simple index
// of them.  For each recipe a link will be displayed which will let you
// trigger the recipe-build.
func IndexHandler(res http.ResponseWriter, req *http.Request) {
	tmpl_src := `
<html>
<head><title>Job Index</title></head>
<body>
<blockquote>
<h1>Job Index</h1>
<ul>
{{range . }}
<li><a href="/run/{{. }}">{{.}}</a></li>
{{end}}
</ul>
</blockquote>
</body>
</html>
`
	//
	// Find all the *.recipe files
	//
	files, err := filepath.Glob("*.recipe")
	if err != nil {
		http.Error(res, "Error finding list of recipes!", http.StatusNotFound)
		return
	}

	//
	// Create a template instance from the source above.
	//
	t := template.Must(template.New("tmpl").Parse(tmpl_src))

	//
	// Execute the template into our buffer.
	//
	buf := &bytes.Buffer{}
	err = t.Execute(buf, files)

	//
	// If there were errors, then show them.
	if err != nil {
		fmt.Fprintf(res, err.Error())
		return
	}

	//
	// Otherwise output the rendered result.
	//
	buf.WriteTo(res)

}

// StartHandler launches a job, and sends the output to the client.
// The job is long-running so we use a streaming writer.
//
// We inject some javascript into the output to follow the output,
// and terminate that on completion.
//
func StartHandler(res http.ResponseWriter, req *http.Request) {

	//
	// Get the job we're going to execute
	//
	vars := mux.Vars(req)
	job := vars["job"]

	//
	// If the job is empty then we abort
	//
	if len(job) < 1 {
		http.Error(res, "Missing 'job'", http.StatusNotFound)
		return
	}

	//
	// Set a sane output type, since it might be non-obvious.
	//
	res.Header().Set("Content-Type", "text/html")

	//
	// Write out a simple header which wraps our output so that
	// it looks pretty - and also handles the auto-scroll behaviour.
	//
	res.Write([]byte(
		`<!DOCTYPE html>
<html lang="en">
 <head>
  <title>Job Output</title>
  <script type="text/javascript">
     var myTimer;

     // scroll to bottom of document every 0.5 seconds
     function enableScroll() {
        myTimer = setInterval(function(){ window.scrollTo(0,document.body.scrollHeight); }, 500);
     }

    // cancel the scrolling
    function stopScroll() {
        clearInterval(myTimer);
    }

    // start now
    enableScroll();
  </script>
 </head>
 <body>
  <pre>`))

	cmd := exec.Command("./thyme", "--recipe", job, "--verbose")
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter
	go writeCmdOutput(res, pipeReader)
	cmd.Run()
	pipeWriter.Close()

	//
	// Terminate
	//
	res.Write([]byte(`</pre>
<script type="text/javascript">
    stopScroll();
    alert("Build Complete!");
</script>
</html>`))

}

// writeCmdOutput reads 255 bytes of output from the command, and streams
// it to the remote client.
func writeCmdOutput(res http.ResponseWriter, pipeReader *io.PipeReader) {
	buffer := make([]byte, 255)
	for {
		n, err := pipeReader.Read(buffer)
		if err != nil {
			pipeReader.Close()
			break
		}

		data := buffer[0:n]
		res.Write(data)
		if f, ok := res.(http.Flusher); ok {
			f.Flush()
		}
		//reset buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

// Entry-point
func main() {

	//
	// Create a new router and our route-mappings.
	//
	router := mux.NewRouter()

	//
	// API end-points
	//
	router.HandleFunc("/", IndexHandler).Methods("GET")
	router.HandleFunc("/run/{job}", StartHandler).Methods("GET")

	//
	// Bind the router.
	//
	http.Handle("/", router)

	//
	// Show where we'll bind
	//
	bindHost := "127.0.0.1"
	bindPort := 8080
	bind := fmt.Sprintf("%s:%d", bindHost, bindPort)
	fmt.Printf("Launching the server on http://%s\n", bind)

	//
	// Wire up logging.
	//
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)

	//
	// We want to make sure we handle timeouts effectively by using
	// a non-default http-server
	//
	srv := &http.Server{
		Addr:        bind,
		Handler:     loggedRouter,
		ReadTimeout: 5 * time.Second,

		// The write-timeout here is huge because we want to
		// ensure that our build-process doesn't get interrupted
		WriteTimeout: 3600 * time.Second,
	}

	//
	// Launch the server.
	//
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Printf("\nError: %s\n", err.Error())
	}
}
