package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"xiangqi-lab/internal/records"
)

func (s *Server) importRecords(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxUploadBytes)
	request, err := decodeImportRequest(r)
	if err != nil {
		s.writeError(w, r, fmt.Errorf("invalid import request: %w", err))
		return
	}
	writeJSON(w, http.StatusCreated, s.records.Import(request))
}

func (s *Server) getImport(w http.ResponseWriter, r *http.Request) {
	batch, err := s.records.GetImport(strings.TrimPrefix(r.PathValue("rest"), "imports/"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func (s *Server) listRecords(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"items": s.records.List()})
}

func (s *Server) getRecord(w http.ResponseWriter, r *http.Request) {
	record, err := s.records.Get(r.PathValue("rest"))
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) getRecordMoves(w http.ResponseWriter, r *http.Request) {
	recordID := strings.TrimSuffix(r.PathValue("rest"), "/moves")
	record, err := s.records.Get(recordID)
	if err != nil {
		s.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"recordId": record.ID, "items": record.Moves})
}

func (s *Server) routeRecordsGet(w http.ResponseWriter, r *http.Request) {
	rest := strings.Trim(r.PathValue("rest"), "/")
	switch {
	case strings.HasPrefix(rest, "imports/") && strings.Count(rest, "/") == 1:
		s.getImport(w, r)
	case strings.HasSuffix(rest, "/moves") && strings.Count(rest, "/") == 1:
		s.getRecordMoves(w, r)
	case rest != "" && !strings.Contains(rest, "/"):
		s.getRecord(w, r)
	default:
		s.writeError(w, r, records.ErrNotFound)
	}
}

func (s *Server) deleteRecord(w http.ResponseWriter, r *http.Request) {
	if err := s.records.Delete(r.PathValue("id")); err != nil {
		s.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeImportRequest(r *http.Request) (records.ImportRequest, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			return records.ImportRequest{}, err
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			return records.ImportRequest{}, err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return records.ImportRequest{}, err
		}
		format := r.FormValue("format")
		if format == "" {
			format = extensionFormat(header)
		}
		return records.ImportRequest{
			Name: r.FormValue("name"), Format: format, Content: string(content),
			InitialFEN: r.FormValue("initialFen"), Result: r.FormValue("result"),
		}, nil
	}
	var request records.ImportRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return records.ImportRequest{}, err
	}
	return request, nil
}

func extensionFormat(header *multipart.FileHeader) string {
	name := strings.ToLower(header.Filename)
	switch {
	case strings.HasSuffix(name, ".json"):
		return "json"
	case strings.HasSuffix(name, ".pgn"):
		return "pgn"
	default:
		return "iccs"
	}
}
