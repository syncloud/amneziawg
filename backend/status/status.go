// Package status exposes live info about the running awg interface —
// peer handshake times, RX/TX bytes, server listen port for the
// "forward this on your router" reminder in the dashboard.
package status

import (
	"encoding/json"
	"net/http"

	"backend/awg"
	"backend/config"
)

type Service struct {
	AWG    *awg.Client
	Config *config.Config
}

type Response struct {
	ListenPort int               `json:"listen_port"`
	AppDomain  string            `json:"app_domain"`
	Peers      []awg.PeerStatus `json:"peers"`
}

func (s *Service) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/status", s.handleStatus)
}

func (s *Service) handleStatus(w http.ResponseWriter, _ *http.Request) {
	peers, err := s.AWG.Dump()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Response{
		ListenPort: s.Config.ListenPort,
		AppDomain:  s.Config.AppDomain,
		Peers:      peers,
	})
}
