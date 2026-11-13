package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"onestarlab/backend/internal/store"
)

// Server 承载 HTTP 路由与静态资源。
type Server struct {
	store      *store.Store
	webDir     string
	apiToken   string
	corsOrigin string
}

func NewServer(st *store.Store, webDir, apiToken, corsOrigin string) *Server {
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	return &Server{store: st, webDir: webDir, apiToken: apiToken, corsOrigin: corsOrigin}
}

// Handler 构建完整路由。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/bootstrap", s.handleBootstrap)

	mux.HandleFunc("GET /api/v1/products", s.handleListProducts)
	mux.HandleFunc("POST /api/v1/products", s.handleCreateProduct)
	mux.HandleFunc("GET /api/v1/products/sku/{sku}", s.handleFindProductBySKU)
	mux.HandleFunc("PUT /api/v1/products/{id}", s.handleUpdateProduct)
	mux.HandleFunc("DELETE /api/v1/products/{id}", s.handleDeleteProduct)
	mux.HandleFunc("POST /api/v1/sku", s.handleGenerateSKU)

	mux.HandleFunc("GET /api/v1/categories", s.handleListCategories)
	mux.HandleFunc("POST /api/v1/categories", s.handleCreateCategory)
	mux.HandleFunc("DELETE /api/v1/categories/{id}", s.handleDeleteCategory)

	mux.HandleFunc("GET /api/v1/orders", s.handleListOrders)
	mux.HandleFunc("POST /api/v1/orders", s.handleCreateOrder)
	mux.HandleFunc("PUT /api/v1/orders/{id}", s.handleUpdateOrder)
	mux.HandleFunc("DELETE /api/v1/orders/{id}", s.handleDeleteOrder)

	mux.HandleFunc("GET /api/v1/pricing-rule", s.handleGetPricingRule)
	mux.HandleFunc("PUT /api/v1/pricing-rule", s.handleSavePricingRule)

	mux.HandleFunc("GET /api/v1/inventory/transactions", s.handleListInventory)
	mux.HandleFunc("POST /api/v1/inventory/transactions", s.handleCreateInventory)
	mux.HandleFunc("DELETE /api/v1/inventory/transactions/{id}", s.handleDeleteInventory)
	mux.HandleFunc("POST /api/v1/inventory/process", s.handleProcessInventory)
	mux.HandleFunc("POST /api/v1/inventory/scan", s.handleScanBatch)

	mux.HandleFunc("GET /api/v1/lab/inspirations", s.handleListInspirations)
	mux.HandleFunc("POST /api/v1/lab/inspirations", s.handleCreateInspiration)
	mux.HandleFunc("PUT /api/v1/lab/inspirations/{id}", s.handleUpdateInspiration)
	mux.HandleFunc("DELETE /api/v1/lab/inspirations/{id}", s.handleDeleteInspiration)

	mux.HandleFunc("GET /api/v1/lab/projects", s.handleListLabProjects)
	mux.HandleFunc("POST /api/v1/lab/projects", s.handleCreateLabProject)
	mux.HandleFunc("PUT /api/v1/lab/projects/{id}", s.handleUpdateLabProject)
	mux.HandleFunc("DELETE /api/v1/lab/projects/{id}", s.handleDeleteLabProject)
	mux.HandleFunc("PATCH /api/v1/lab/projects/{id}/stage", s.handleUpdateLabStage)
	mux.HandleFunc("POST /api/v1/lab/projects/{id}/proofs", s.handleAddProof)
	mux.HandleFunc("POST /api/v1/lab/projects/{id}/convert", s.handleConvertProject)

	mux.HandleFunc("GET /api/v1/drawings", s.handleListDrawings)
	mux.HandleFunc("POST /api/v1/drawings", s.handleCreateDrawing)
	mux.HandleFunc("PUT /api/v1/drawings/{id}", s.handleUpdateDrawing)
	mux.HandleFunc("DELETE /api/v1/drawings/{id}", s.handleDeleteDrawing)

	mux.HandleFunc("GET /api/v1/expenses", s.handleListExpenses)
	mux.HandleFunc("POST /api/v1/expenses", s.handleCreateExpense)
	mux.HandleFunc("PUT /api/v1/expenses/{id}", s.handleUpdateExpense)
	mux.HandleFunc("DELETE /api/v1/expenses/{id}", s.handleDeleteExpense)

	mux.HandleFunc("GET /api/v1/markets", s.handleListMarkets)
	mux.HandleFunc("POST /api/v1/markets", s.handleCreateMarket)
	mux.HandleFunc("PUT /api/v1/markets/{id}", s.handleUpdateMarket)
	mux.HandleFunc("DELETE /api/v1/markets/{id}", s.handleDeleteMarket)

	mux.HandleFunc("GET /api/v1/settings/online-income", s.handleGetOnlineIncome)
	mux.HandleFunc("PUT /api/v1/settings/online-income", s.handleSaveOnlineIncome)

	mux.HandleFunc("GET /api/v1/planner-templates", s.handleListPlannerTemplates)
	mux.HandleFunc("POST /api/v1/planner-templates", s.handleCreatePlannerTemplate)
	mux.HandleFunc("PUT /api/v1/planner-templates/{id}", s.handleUpdatePlannerTemplate)
	mux.HandleFunc("DELETE /api/v1/planner-templates/{id}", s.handleDeletePlannerTemplate)

	mux.HandleFunc("GET /api/v1/export", s.handleExport)
	mux.HandleFunc("POST /api/v1/import", s.handleImport)

	mux.HandleFunc("/", s.handleStatic)

	return s.withMiddleware(mux)
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if origin := r.Header.Get("Origin"); origin != "" {
			allowed := s.corsOrigin
			if allowed == "*" || origin == allowed || containsOrigin(allowed, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "600")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if s.apiToken != "" && strings.HasPrefix(r.URL.Path, "/api/") {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer "+s.apiToken {
				writeError(w, http.StatusUnauthorized, "未授权")
				return
			}
		}

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start).Round(time.Millisecond))
	})
}

func containsOrigin(list, origin string) bool {
	for _, item := range strings.Split(list, ",") {
		if strings.TrimSpace(item) == origin {
			return true
		}
	}
	return false
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "time": time.Now().UTC().Format(time.RFC3339)})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("encode response: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// fail 将业务错误映射为合适的 HTTP 状态码。
func fail(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "记录不存在")
		return
	}
	msg := err.Error()
	if strings.Contains(msg, "未找到") || strings.Contains(msg, "库存不足") ||
		strings.Contains(msg, "不可删除") || strings.Contains(msg, "需为") ||
		strings.Contains(msg, "已转为") || strings.Contains(msg, "没有扫描到") {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "服务器内部错误")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误: "+err.Error())
		return false
	}
	return true
}

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeError(w, http.StatusNotFound, "接口不存在")
		return
	}
	if s.webDir == "" {
		writeError(w, http.StatusNotFound, "前端静态资源未配置")
		return
	}

	clean := filepath.Clean("/" + r.URL.Path)
	filePath := filepath.Join(s.webDir, clean)
	info, err := os.Stat(filePath)
	if err == nil && !info.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	// SPA 路由回退到 index.html。
	indexPath := filepath.Join(s.webDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		writeError(w, http.StatusNotFound, "前端静态资源未找到")
		return
	}
	http.ServeFile(w, r, indexPath)
}
