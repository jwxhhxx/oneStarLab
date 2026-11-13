package api

import (
	"net/http"

	"onestarlab/backend/internal/model"
)

func (s *Server) handleListInspirations(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListInspirations()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateInspiration(w http.ResponseWriter, r *http.Request) {
	var input model.InspirationInput
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := s.store.CreateInspiration(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleUpdateInspiration(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.InspirationInput
	if !decodeJSON(w, r, &input) {
		return
	}
	item, err := s.store.UpdateInspiration(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleDeleteInspiration(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteInspiration(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListLabProjects(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListLabProjects()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateLabProject(w http.ResponseWriter, r *http.Request) {
	var input model.LabProjectInput
	if !decodeJSON(w, r, &input) {
		return
	}
	project, err := s.store.CreateLabProject(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, project)
}

func (s *Server) handleUpdateLabProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.LabProjectInput
	if !decodeJSON(w, r, &input) {
		return
	}
	project, err := s.store.UpdateLabProject(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleDeleteLabProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteLabProject(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleUpdateLabStage(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input struct {
		Stage string `json:"stage"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	project, err := s.store.UpdateLabProjectStage(id, input.Stage)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleAddProof(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.ProofRecordInput
	if !decodeJSON(w, r, &input) {
		return
	}
	project, err := s.store.AddProofRecord(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleConvertProject(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	product, err := s.store.ConvertLabProjectToProduct(id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (s *Server) handleListDrawings(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListDrawingInspirations()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateDrawing(w http.ResponseWriter, r *http.Request) {
	var input model.DrawingInspirationRecord
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.CreateDrawingInspiration(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s *Server) handleUpdateDrawing(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var patch model.DrawingInspirationRecord
	if !decodeJSON(w, r, &patch) {
		return
	}
	if err := s.store.UpdateDrawingInspiration(id, patch); err != nil {
		fail(w, err)
		return
	}
	record, err := s.store.GetDrawingInspiration(id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (s *Server) handleDeleteDrawing(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteDrawingInspiration(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
