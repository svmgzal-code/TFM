package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"tfm/backend/internal/protocol"
	"tfm/backend/internal/radar"
)

type connectRequest struct {
	Host        string `json:"host"`
	CommandPort int    `json:"commandPort"`
	StreamPort  int    `json:"streamPort"`
}

type statusResponse struct {
	Connection  string `json:"connection"`
	Recording   bool   `json:"recording"`
	Host        string `json:"host"`
	CommandPort int    `json:"commandPort"`
	StreamPort  int    `json:"streamPort"`
}

type infoResponse struct {
	FirmwareVersion      string `json:"firmwareVersion"`
	FirmwareRevision     uint32 `json:"firmwareRevision"`
	SntBoardVersion      string `json:"sntBoardVersion"`
	BasebandBoardVersion string `json:"basebandBoardVersion"`
	FrontendBoardVersion string `json:"frontendBoardVersion"`
	AvailableChannels    uint16 `json:"availableChannels"`
	AvailableAlgorithms  uint16 `json:"availableAlgorithms"`
	RadarHardwareID      uint16 `json:"radarHardwareId"`
	RadarNumber          uint32 `json:"radarNumber"`
	FlashingDate         string `json:"flashingDate"`
	PhaseOffsetDeg       int32  `json:"phaseOffsetDeg"`
}

type paramsResponse struct {
	Band           uint8  `json:"band"`
	TRampMs        uint8  `json:"tRampMs"`
	ZeroPad        uint8  `json:"zeroPad"`
	FftDataType    uint8  `json:"fftDataType"`
	FrontendEn     bool   `json:"frontendEn"`
	PowerSaveEn    bool   `json:"powerSaveEn"`
	Norm           bool   `json:"norm"`
	ActChannels    uint16 `json:"actChannels"`
	FreqPoints     uint16 `json:"freqPoints"`
	MinFreqMHz     uint16 `json:"minFreqMHz"`
	ManualBWMHz    uint16 `json:"manualBWMHz"`
	Attenuation    uint8  `json:"attenuation"`
	RangeBinUm     uint32 `json:"rangeBinUm"`
	DopplerBinUmS  uint32 `json:"dopplerBinUmS"`
	FrequencyBinHz uint32 `json:"frequencyBinHz"`
}

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	var req connectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Host == "" || req.CommandPort <= 0 || req.StreamPort <= 0 {
		writeError(w, http.StatusBadRequest, "host, commandPort and streamPort are required")
		return
	}

	if err := s.session.Connect(r.Context(), req.Host, req.CommandPort, req.StreamPort); err != nil {
		writeError(w, mapError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.status())
}

func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if err := s.session.Disconnect(r.Context()); err != nil {
		writeError(w, mapError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.status())
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.status())
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	info, ok := s.session.Info()
	if !ok {
		writeError(w, http.StatusConflict, "not connected")
		return
	}
	writeJSON(w, http.StatusOK, toInfoResponse(info))
}

func (s *Server) handleParams(w http.ResponseWriter, r *http.Request) {
	params, ok := s.session.Params()
	if !ok {
		writeError(w, http.StatusConflict, "not connected")
		return
	}
	writeJSON(w, http.StatusOK, toParamsResponse(params))
}

func (s *Server) status() statusResponse {
	st := s.session.Status()
	return statusResponse{
		Connection:  st.State.String(),
		Recording:   false,
		Host:        st.Host,
		CommandPort: st.CommandPort,
		StreamPort:  st.StreamPort,
	}
}

func toInfoResponse(i protocol.Info) infoResponse {
	return infoResponse{
		FirmwareVersion:      protocol.FirmwareVersionString(i.FirmwareVersion),
		FirmwareRevision:     i.FirmwareRevision,
		SntBoardVersion:      protocol.Hex32(i.SntBoardVersion),
		BasebandBoardVersion: protocol.Hex32(i.BasebandBoardVersion),
		FrontendBoardVersion: protocol.Hex32(i.FrontendBoardVersion),
		AvailableChannels:    i.AvailableChannels,
		AvailableAlgorithms:  i.AvailableAlgorithms,
		RadarHardwareID:      i.RadarHardwareID,
		RadarNumber:          i.RadarNumber,
		FlashingDate:         protocol.FlashingDateString(i.FlashingDate),
		PhaseOffsetDeg:       i.PhaseOffset,
	}
}

func toParamsResponse(p protocol.Params) paramsResponse {
	return paramsResponse{
		Band:           p.Band,
		TRampMs:        p.TRamp,
		ZeroPad:        p.ZeroPad,
		FftDataType:    p.FftDataType,
		FrontendEn:     p.FrontendEn != 0,
		PowerSaveEn:    p.PowerSaveEn != 0,
		Norm:           p.Norm != 0,
		ActChannels:    p.ActChannels,
		FreqPoints:     p.FreqPoints,
		MinFreqMHz:     p.MinFreq,
		ManualBWMHz:    p.ManualBW,
		Attenuation:    p.Attenuation,
		RangeBinUm:     p.RangeBin,
		DopplerBinUmS:  p.DopplerBin,
		FrequencyBinHz: p.FrequencyBin,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func mapError(err error) int {
	switch {
	case errors.Is(err, radar.ErrAlreadyConnected), errors.Is(err, radar.ErrNotConnected):
		return http.StatusConflict
	default:
		return http.StatusBadGateway
	}
}
