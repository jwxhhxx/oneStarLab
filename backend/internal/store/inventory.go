package store

import (
	"database/sql"
	"fmt"

	"onestarlab/backend/internal/model"
)

func (s *Store) ListInventoryTransactions() ([]model.InventoryTransaction, error) {
	rows, err := s.db.Query(`SELECT id, product_id, barcode, type, quantity, note, created_at
		FROM inventory_transactions ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.InventoryTransaction, 0)
	for rows.Next() {
		var t model.InventoryTransaction
		if err := rows.Scan(&t.ID, &t.ProductID, &t.Barcode, &t.Type, &t.Quantity, &t.Note, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Store) AddInventoryTransaction(input model.InventoryTransactionInput) (model.InventoryTransaction, error) {
	createdAt := nowISO()
	res, err := s.db.Exec(`INSERT INTO inventory_transactions (product_id, barcode, type, quantity, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, input.ProductID, input.Barcode, input.Type, input.Quantity, input.Note, createdAt)
	if err != nil {
		return model.InventoryTransaction{}, err
	}
	id, _ := res.LastInsertId()
	return model.InventoryTransaction{
		ID: id, ProductID: input.ProductID, Barcode: input.Barcode, Type: input.Type,
		Quantity: input.Quantity, Note: input.Note, CreatedAt: createdAt,
	}, nil
}

func (s *Store) DeleteInventoryTransaction(id int64) error {
	res, err := s.db.Exec(`DELETE FROM inventory_transactions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ProcessInventoryTransaction(input model.InventoryProcessInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var product model.Product
	row := tx.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, input.ProductID)
	if err := scanProductInto(row, &product); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("未找到商品")
		}
		return err
	}

	newStock := product.Stock + input.Quantity
	if input.Type == "out" {
		newStock = product.Stock - input.Quantity
		if newStock < 0 {
			return fmt.Errorf("库存不足")
		}
		customer := "扫码出库"
		if input.Note != "" {
			customer = "扫码:" + input.Note
		}
		order := buildSingleOrder(product, model.NewOrderInput{
			ProductID: input.ProductID, CustomerName: customer, Channel: "扫码", Quantity: input.Quantity,
		}, orderNo(), nowISO(), "paid")
		if _, err := insertOrder(tx, order); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(`UPDATE products SET stock = ? WHERE id = ?`, newStock, product.ID); err != nil {
		return err
	}
	barcode := input.Barcode
	if barcode == "" {
		barcode = product.SKU
	}
	if _, err := tx.Exec(`INSERT INTO inventory_transactions (product_id, barcode, type, quantity, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, input.ProductID, barcode, input.Type, input.Quantity, input.Note, nowISO()); err != nil {
		return err
	}
	return tx.Commit()
}

// ScanBatch 批量扫码出入库：聚合同一商品、校验库存、生成订单与流水。
func (s *Store) ScanBatch(input model.ScanBatchInput) (string, error) {
	if len(input.Items) == 0 {
		return "", fmt.Errorf("没有扫描到任何商品")
	}

	type aggItem struct {
		productID int64
		quantity  int
		barcode   string
	}
	order := make([]int64, 0)
	agg := make(map[int64]*aggItem)
	for _, item := range input.Items {
		if cur, ok := agg[item.ProductID]; ok {
			cur.quantity += item.Quantity
		} else {
			agg[item.ProductID] = &aggItem{productID: item.ProductID, quantity: item.Quantity, barcode: item.Barcode}
			order = append(order, item.ProductID)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	products := make(map[int64]model.Product, len(order))
	for _, pid := range order {
		var p model.Product
		row := tx.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
			sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, pid)
		if err := scanProductInto(row, &p); err != nil {
			if err == sql.ErrNoRows {
				return "", fmt.Errorf("未找到商品 id=%d", pid)
			}
			return "", err
		}
		if input.Type == "out" && p.Stock < agg[pid].quantity {
			return "", fmt.Errorf("商品 %s 库存不足", p.Name)
		}
		products[pid] = p
	}

	items := make([]model.OrderItem, 0, len(order))
	var totalAmount, goodsCost float64
	for _, pid := range order {
		p := products[pid]
		qty := agg[pid].quantity
		itemTotal := round2(p.SalePrice * float64(qty))
		itemCost := round2((p.PurchaseCost + p.PackagingCost) * float64(qty))
		items = append(items, model.OrderItem{
			ProductID: p.ID, ProductName: p.Name, SKU: p.SKU, Quantity: qty,
			SalePrice: p.SalePrice, TotalAmount: itemTotal, TotalCost: itemCost,
		})
		totalAmount += itemTotal
		goodsCost += itemCost
	}

	customer := "扫码出库批量"
	if input.Note != "" {
		customer = "扫码:" + input.Note
	}
	o := model.Order{
		OrderNo:        orderNo(),
		Channel:        "扫码",
		CustomerName:   customer,
		Items:          items,
		TotalAmount:    round2(totalAmount),
		DiscountAmount: 0,
		ShippingCost:   0,
		PlatformFee:    0,
		GoodsCost:      round2(goodsCost),
		NetProfit:      round2(totalAmount - goodsCost),
		CreatedAt:      nowISO(),
		Status:         "paid",
	}
	if _, err := insertOrder(tx, o); err != nil {
		return "", err
	}

	createdAt := nowISO()
	for _, pid := range order {
		p := products[pid]
		qty := agg[pid].quantity
		newStock := p.Stock + qty
		if input.Type == "out" {
			newStock = p.Stock - qty
		}
		if _, err := tx.Exec(`UPDATE products SET stock = ? WHERE id = ?`, newStock, p.ID); err != nil {
			return "", err
		}
		barcode := agg[pid].barcode
		if barcode == "" {
			barcode = p.SKU
		}
		if _, err := tx.Exec(`INSERT INTO inventory_transactions (product_id, barcode, type, quantity, note, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`, p.ID, barcode, input.Type, qty, input.Note, createdAt); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return o.OrderNo, nil
}
