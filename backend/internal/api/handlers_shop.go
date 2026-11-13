package api

import (
	"net/http"

	"onestarlab/backend/internal/model"
)

func (s *Server) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	data, err := s.store.Bootstrap()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (s *Server) handleListProducts(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListProducts()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var input model.ProductInput
	if !decodeJSON(w, r, &input) {
		return
	}
	product, err := s.store.CreateProduct(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

func (s *Server) handleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.ProductInput
	if !decodeJSON(w, r, &input) {
		return
	}
	product, err := s.store.UpdateProduct(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (s *Server) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteProduct(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleFindProductBySKU(w http.ResponseWriter, r *http.Request) {
	product, err := s.store.FindProductBySKU(r.PathValue("sku"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

func (s *Server) handleGenerateSKU(w http.ResponseWriter, r *http.Request) {
	var input model.SKURequest
	if !decodeJSON(w, r, &input) {
		return
	}
	sku, err := s.store.GenerateSKU(input.Category, input.ExcludeProductID, input.CreateIfMissing)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sku": sku})
}

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListCategories()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	category, err := s.store.AddCategory(input.Name)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, category)
}

func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	reassign := r.URL.Query().Get("reassignTo")
	if err := s.store.DeleteCategory(id, reassign); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListOrders(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListOrders()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var input model.NewOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}
	order, err := s.store.CreateOrder(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (s *Server) handleUpdateOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	var input model.NewOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}
	order, err := s.store.UpdateOrder(id, input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) handleDeleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteOrder(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleGetPricingRule(w http.ResponseWriter, r *http.Request) {
	rule, err := s.store.GetPricingRule()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) handleSavePricingRule(w http.ResponseWriter, r *http.Request) {
	var rule model.PricingRule
	if !decodeJSON(w, r, &rule) {
		return
	}
	if err := s.store.SavePricingRule(rule); err != nil {
		fail(w, err)
		return
	}
	rule.ID = 1
	writeJSON(w, http.StatusOK, rule)
}

func (s *Server) handleListInventory(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListInventoryTransactions()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCreateInventory(w http.ResponseWriter, r *http.Request) {
	var input model.InventoryTransactionInput
	if !decodeJSON(w, r, &input) {
		return
	}
	record, err := s.store.AddInventoryTransaction(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, record)
}

func (s *Server) handleDeleteInventory(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的 id")
		return
	}
	if err := s.store.DeleteInventoryTransaction(id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleProcessInventory(w http.ResponseWriter, r *http.Request) {
	var input model.InventoryProcessInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.store.ProcessInventoryTransaction(input); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleScanBatch(w http.ResponseWriter, r *http.Request) {
	var input model.ScanBatchInput
	if !decodeJSON(w, r, &input) {
		return
	}
	no, err := s.store.ScanBatch(input)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"orderNo": no})
}
