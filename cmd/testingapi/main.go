package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:9876")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	mux := http.NewServeMux()
	mux.Handle("/install", http.HandlerFunc(handleInstall))
	mux.Handle("/run", http.HandlerFunc(handleRun))

	srv := &http.Server{
		Handler: mux,
	}
	srv.Serve(l)
}

func handleInstall(w http.ResponseWriter, req *http.Request) {
	cmd := exec.CommandContext(req.Context(), "/opt/moby/install.sh")
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(out), http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

type RunResponse struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Junit  string `json:"junit"`
}

func handleRun(w http.ResponseWriter, req *http.Request) {
	cmd := exec.CommandContext(req.Context(), "/opt/moby/test.sh")
	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)

	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	err := cmd.Run()

	dt, readErr := os.ReadFile("/opt/moby/TestReport-test.sh.xml")

	resp := RunResponse{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
		Junit:  string(dt),
	}

	dt, marshalErr := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	switch {
	case err != nil:
		w.WriteHeader(http.StatusTeapot)
	case marshalErr != nil || readErr != nil:
		w.WriteHeader(http.StatusInternalServerError)
	default:
		w.WriteHeader(http.StatusOK)
	}

	if _, err := w.Write(dt); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write response: ", err)
	}
}
