package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lavavrik/go-sm/protobufs"
	"github.com/lavavrik/go-sm/stats"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// FileListResponse represents the response structure for file listing
type FileListResponse struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// ListFiles handles the file listing endpoint
func ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the path from query parameter, default to current directory
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	// Clean the path to prevent directory traversal attacks
	path = filepath.Clean(path)

	// Read directory contents
	entries, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to our FileInfo structure
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't read info for
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		}
		files = append(files, fileInfo)
	}

	// Create response
	response := FileListResponse{
		Path:  path,
		Files: files,
	}

	// Set content type and encode JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

// HealthCheck provides a simple health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "go-sm file API",
	})
}

// Read load from stats.log file
func OverallLoad(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queryParams := r.URL.Query()

	sinceString := queryParams.Get("since")
	untilString := queryParams.Get("until")
	responseType := queryParams.Get("type")
	if responseType == "" {
		responseType = "json"
	}

	var since, until uint64
	var err error
	if sinceString != "" {
		since, err = strconv.ParseUint(sinceString, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid since parameter: %v", err), http.StatusBadRequest)
			return
		}
	}
	if untilString != "" {
		until, err = strconv.ParseUint(untilString, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid until parameter: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Read the stats.log file
	data, err := os.ReadFile("stats.log")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading stats.log: %v", err), http.StatusInternalServerError)
		return
	}

	dataPoints, err := stats.BytesToDataPoints(data)

	filteredDataPoints := []*protobufs.DataPoint{}

	// Apply filtering based on since and until if provided
	for _, dp := range dataPoints {
		include := true

		if sinceString != "" && dp.Timestamp < since {
			include = false
		}
		if untilString != "" && dp.Timestamp > until {
			include = false
		}

		if include {
			filteredDataPoints = append(filteredDataPoints, dp)
		}
	}

	switch responseType {
	case "proto":
		protoData := stats.DataPointsToBytes(filteredDataPoints)

		// Set content type and write response
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(protoData)))
		w.WriteHeader(http.StatusOK)
		w.Write(protoData)
		return
	case "json":
		// Marshal to JSON for response
		jsonData, err := json.Marshal(filteredDataPoints)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error marshaling to JSON: %v", err), http.StatusInternalServerError)
			return
		}

		// Set content type and write response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(jsonData)))
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}
}

// RegisterHandlers registers all API handlers with the default HTTP mux
func RegisterHandlers() {
	http.HandleFunc("/health", HealthCheck)
	http.HandleFunc("/files", ListFiles)
	http.HandleFunc("/stats", OverallLoad)
}
