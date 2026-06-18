package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type serverConfig struct {
	Bind     string
	Port     int
	FileRoot string
}

type dirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	Path string `json:"path"`
}

type uploadResponse struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	cfg := parseFlags()
	if err := os.MkdirAll(cfg.FileRoot, 0o755); err != nil {
		log.Fatalf("create file root directory: %v", err)
	}

	logInfo("http4dev server started")
	logInfo("  FileRoot: %s", cfg.FileRoot)
	logInfo("  Listen: %s:%d", cfg.Bind, cfg.Port)
	logInfo("  Endpoints:")
	logInfo("    GET  /<path>        Download file")
	logInfo("    PUT  /<path>        Raw upload")
	logInfo("    POST /upload        Multipart form upload")
	logInfo("    GET  /list?path=    List directory")
	logInfo("")

	router := newRouter(cfg.FileRoot)
	addr := fmt.Sprintf("%s:%d", cfg.Bind, cfg.Port)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func parseFlags() serverConfig {
	var cfg serverConfig
	flag.IntVar(&cfg.Port, "port", 8080, "Port to listen on")
	flag.IntVar(&cfg.Port, "p", 8080, "Port to listen on")
	flag.StringVar(&cfg.FileRoot, "root", "./file_root", "File root directory for file storage")
	flag.StringVar(&cfg.FileRoot, "r", "./file_root", "File root directory for file storage")
	flag.StringVar(&cfg.Bind, "bind", "0.0.0.0", "Address to bind to")
	flag.StringVar(&cfg.Bind, "b", "0.0.0.0", "Address to bind to")
	flag.Parse()

	absRoot, err := filepath.Abs(cfg.FileRoot)
	if err != nil {
		log.Fatalf("resolve file root directory: %v", err)
	}
	cfg.FileRoot = filepath.Clean(absRoot)
	return cfg
}

func newRouter(fileRoot string) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/list", func(c *gin.Context) {
		handleList(c, fileRoot)
	})
	router.POST("/upload", func(c *gin.Context) {
		handleUpload(c, fileRoot)
	})

	router.NoRoute(func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			handleGet(c, fileRoot)
		case http.MethodPut:
			handlePut(c, fileRoot)
		default:
			c.JSON(http.StatusNotFound, errorResponse{Error: "Not found"})
		}
	})

	return router
}

func handleList(c *gin.Context, fileRoot string) {
	dirPath := c.Query("path")
	fsPath, ok := safeJoin(fileRoot, dirPath)
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Path traversal rejected"})
		return
	}

	info, err := os.Stat(fsPath)
	if err != nil || !info.IsDir() {
		c.JSON(http.StatusNotFound, errorResponse{Error: fmt.Sprintf("Directory not found: %s", dirPath)})
		return
	}

	entries, err := listDirectory(fsPath, fileRoot)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func handleUpload(c *gin.Context, fileRoot string) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Missing 'file' field in form data"})
		return
	}
	if fileHeader.Filename == "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Empty filename"})
		return
	}

	pathField := c.PostForm("path")
	filename := sanitizeFilename(fileHeader.Filename)
	targetPath, ok := safeJoin(fileRoot, pathField, filename)
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Path traversal rejected"})
		return
	}

	size, err := saveMultipartFile(fileHeader, targetPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	relPath, err := filepath.Rel(fileRoot, targetPath)
	if err != nil {
		relPath = targetPath
	}
	relPath = filepath.ToSlash(relPath)
	logInfo("POST uploaded: %s (%d bytes)", relPath, size)
	c.JSON(http.StatusOK, uploadResponse{Status: "ok", Path: relPath, Size: size})
}

func handlePut(c *gin.Context, fileRoot string) {
	urlPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	fsPath, ok := safeJoin(fileRoot, urlPath)
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Path traversal rejected"})
		return
	}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Empty request body"})
		return
	}

	if err := os.MkdirAll(filepath.Dir(fsPath), 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	if err := os.WriteFile(fsPath, data, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}

	logInfo("PUT uploaded: %s (%d bytes)", urlPath, len(data))
	c.JSON(http.StatusOK, uploadResponse{Status: "ok", Path: urlPath, Size: int64(len(data))})
}

func handleGet(c *gin.Context, fileRoot string) {
	urlPath := strings.TrimPrefix(c.Request.URL.Path, "/")
	fsPath, ok := safeJoin(fileRoot, urlPath)
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "Path traversal rejected"})
		return
	}

	info, err := os.Stat(fsPath)
	if err != nil || info.IsDir() {
		c.JSON(http.StatusNotFound, errorResponse{Error: fmt.Sprintf("File not found: %s", urlPath)})
		return
	}

	logInfo("Downloaded: %s (%d bytes)", urlPath, info.Size())
	c.FileAttachment(fsPath, filepath.Base(fsPath))
}

func safeJoin(fileRoot string, paths ...string) (string, bool) {
	parts := make([]string, 0, len(paths)+1)
	parts = append(parts, fileRoot)
	for _, p := range paths {
		p = strings.ReplaceAll(p, "\\", "/")
		p = strings.TrimLeft(p, "/")
		if p == "" {
			continue
		}
		elem := filepath.FromSlash(p)
		if filepath.IsAbs(elem) || filepath.VolumeName(elem) != "" {
			return "", false
		}
		parts = append(parts, elem)
	}

	combined := filepath.Clean(filepath.Join(parts...))
	rel, err := filepath.Rel(fileRoot, combined)
	if err != nil {
		return "", false
	}
	if rel == "." {
		return combined, true
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", false
	}
	return combined, true
}

func sanitizeFilename(filename string) string {
	filename = filepath.Base(strings.ReplaceAll(filename, "\\", "/"))
	var b strings.Builder
	for _, r := range filename {
		switch r {
		case '<', '>', ':', '"', '|', '?', '*':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	filename = b.String()
	if filename == "" || filename == "." || filename == string(filepath.Separator) {
		return "unnamed_file"
	}
	return filename
}

func listDirectory(dirPath, fileRoot string) ([]dirEntry, error) {
	items, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name() < items[j].Name() })

	entries := make([]dirEntry, 0, len(items))
	for _, item := range items {
		fullPath := filepath.Join(dirPath, item.Name())
		info, err := item.Info()
		if err != nil {
			return nil, err
		}
		relPath, err := filepath.Rel(fileRoot, fullPath)
		if err != nil {
			return nil, err
		}
		entry := dirEntry{
			Name: item.Name(),
			Size: info.Size(),
			Path: filepath.ToSlash(relPath),
		}
		if item.IsDir() {
			entry.Type = "dir"
			entry.Size = 0
		} else {
			entry.Type = "file"
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func saveMultipartFile(fileHeader *multipart.FileHeader, targetPath string) (int64, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return 0, err
	}
	out, err := os.Create(targetPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	size, err := io.Copy(out, file)
	if err != nil {
		return size, err
	}
	return size, nil
}

func logInfo(format string, args ...any) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	fmt.Printf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), message)
}
