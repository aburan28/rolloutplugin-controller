package controller

import (
	"net/http"
	"net/http/pprof"
)

func NewPProfServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline/", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile/", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol/", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace/", pprof.Trace)

	return mux
}
