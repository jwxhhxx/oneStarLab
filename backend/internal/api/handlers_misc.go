package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"onestarlab/backend/internal/model"
	"onestarlab/backend/internal/store"
)

func (s *Server) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListExpenses()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	var input model.ExpenseInput
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.CreateExpense(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s *Server) handleUpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.ExpenseInput
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.UpdateExpense(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteExpense(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListMarkets(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListMarkets()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateMarket(w http.ResponseWriter, r *http.Request) {
	var input model.MarketInput
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.CreateMarket(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s *Server) handleUpdateMarket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.MarketInput
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.UpdateMarket(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleDeleteMarket(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteMarket(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleGetOnlineIncome(w http.ResponseWriter, r *http.Request) {
	amount, updatedAt, err := s.store.GetOnlineIncome()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"onlineIncome": amount, "updatedAt": updatedAt})
}

func (s *Server) handleSaveOnlineIncome(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Amount float64 `json:"amount"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.store.SaveOnlineIncome(input.Amount); err != nil {
		fail(w, err)
		return
	}
	amount, updatedAt, err := s.store.GetOnlineIncome()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"onlineIncome": amount, "updatedAt": updatedAt})
}

func (s *Server) handleListPlannerTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListPlannerTemplates()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreatePlannerTemplate(w http.ResponseWriter, r *http.Request) {
	var input model.PlannerTemplateInput
	if !decodeJSON(w, r, &input) {
		return
	}
	template, err := s.store.SavePlannerTemplate(input, 0)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, template)
}

func (s *Server) handleUpdatePlannerTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.PlannerTemplateInput
	if !decodeJSON(w, r, &input) {
		return
	}
	template, err := s.store.SavePlannerTemplate(input, id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, template)
}

func (s *Server) handleDeletePlannerTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeletePlannerTemplate(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.store.Snapshot(time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="onestarlab-backup.json"`)
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 64<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "读取导入文件失败："+err.Error())
		return
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		writeError(w, http.StatusBadRequest, "导入文件为空")
		return
	}

	var snapshot model.Snapshot
	source := "snapshot"
	var probe struct {
		FormatName string `json:"formatName"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		writeError(w, http.StatusBadRequest, "导入文件不是有效的 JSON")
		return
	}
	if probe.FormatName != "" {
		// 旧版纯前端（Dexie/IndexedDB）的导出文件。
		snapshot, err = store.ParseLegacyExport(body)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		source = "dexie"
	} else {
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&snapshot); err != nil {
			writeError(w, http.StatusBadRequest, "导入文件格式错误："+err.Error())
			return
		}
	}

	if err := s.store.ImportSnapshot(snapshot); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "source": source})
}
