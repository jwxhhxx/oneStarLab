package store

import (
	"database/sql"
	"fmt"
	"time"

	"onestarlab/backend/internal/model"
)

func (s *Store) ListOrders() ([]model.Order, error) {
	rows, err := s.db.Query(`SELECT id, order_no, channel, customer_name, items_json, total_amount,
		discount_amount, shipping_cost, platform_fee, goods_cost, net_profit, created_at, status
		FROM orders ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		var itemsJSON string
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.Channel, &o.CustomerName, &itemsJSON, &o.TotalAmount,
			&o.DiscountAmount, &o.ShippingCost, &o.PlatformFee, &o.GoodsCost, &o.NetProfit,
			&o.CreatedAt, &o.Status); err != nil {
			return nil, err
		}
		if err := unmarshalItems(itemsJSON, &o.Items); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (s *Store) GetOrder(id int64) (model.Order, error) {
	var o model.Order
	var itemsJSON string
	err := s.db.QueryRow(`SELECT id, order_no, channel, customer_name, items_json, total_amount,
		discount_amount, shipping_cost, platform_fee, goods_cost, net_profit, created_at, status
		FROM orders WHERE id = ?`, id).Scan(&o.ID, &o.OrderNo, &o.Channel, &o.CustomerName,
		&itemsJSON, &o.TotalAmount, &o.DiscountAmount, &o.ShippingCost, &o.PlatformFee,
		&o.GoodsCost, &o.NetProfit, &o.CreatedAt, &o.Status)
	if err == sql.ErrNoRows {
		return model.Order{}, ErrNotFound
	}
	if err != nil {
		return model.Order{}, err
	}
	if err := unmarshalItems(itemsJSON, &o.Items); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func orderNo() string {
	return "ORD-" + time.Now().Format("20060102-150405")
}

// buildSingleOrder 复刻前端 buildOrderPayload 的口径。
func buildSingleOrder(product model.Product, input model.NewOrderInput, no, createdAt, status string) model.Order {
	originalAmount := round2(product.SalePrice * float64(input.Quantity))
	totalAmount := round2(mathMax(0, originalAmount-input.DiscountAmount))
	platformFee := round2(totalAmount * input.PlatformFeeRate)
	goodsCost := round2((product.PurchaseCost + product.PackagingCost) * float64(input.Quantity))
	netProfit := round2(totalAmount - input.ShippingCost - platformFee - goodsCost)

	return model.Order{
		OrderNo:      no,
		Channel:      input.Channel,
		CustomerName: input.CustomerName,
		Items: []model.OrderItem{{
			ProductID:   product.ID,
			ProductName: product.Name,
			SKU:         product.SKU,
			Quantity:    input.Quantity,
			SalePrice:   product.SalePrice,
			TotalAmount: totalAmount,
			TotalCost:   goodsCost,
		}},
		TotalAmount:    totalAmount,
		DiscountAmount: input.DiscountAmount,
		ShippingCost:   input.ShippingCost,
		PlatformFee:    platformFee,
		GoodsCost:      goodsCost,
		NetProfit:      netProfit,
		CreatedAt:      createdAt,
		Status:         status,
	}
}

func mathMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func insertOrder(tx *sql.Tx, o model.Order) (int64, error) {
	itemsJSON, err := marshal(o.Items)
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(`INSERT INTO orders (order_no, channel, customer_name, items_json, total_amount,
		discount_amount, shipping_cost, platform_fee, goods_cost, net_profit, created_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		o.OrderNo, o.Channel, o.CustomerName, itemsJSON, o.TotalAmount, o.DiscountAmount,
		o.ShippingCost, o.PlatformFee, o.GoodsCost, o.NetProfit, o.CreatedAt, o.Status)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) CreateOrder(input model.NewOrderInput) (model.Order, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback()

	var product model.Product
	row := tx.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, input.ProductID)
	if err := scanProductInto(row, &product); err != nil {
		if err == sql.ErrNoRows {
			return model.Order{}, fmt.Errorf("未找到商品")
		}
		return model.Order{}, err
	}
	if product.Stock < input.Quantity {
		return model.Order{}, fmt.Errorf("库存不足，请先补货")
	}

	o := buildSingleOrder(product, input, orderNo(), nowISO(), "paid")
	id, err := insertOrder(tx, o)
	if err != nil {
		return model.Order{}, err
	}
	if _, err := tx.Exec(`UPDATE products SET stock = ? WHERE id = ?`, product.Stock-input.Quantity, product.ID); err != nil {
		return model.Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Order{}, err
	}
	o.ID = id
	return o, nil
}

func (s *Store) UpdateOrder(id int64, input model.NewOrderInput) (model.Order, error) {
	existing, err := s.GetOrder(id)
	if err != nil {
		return model.Order{}, err
	}
	if len(existing.Items) == 0 {
		return model.Order{}, fmt.Errorf("订单缺少商品信息")
	}
	previousItem := existing.Items[0]

	tx, err := s.db.Begin()
	if err != nil {
		return model.Order{}, err
	}
	defer tx.Rollback()

	var previousProduct, nextProduct model.Product
	row := tx.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, previousItem.ProductID)
	if err := scanProductInto(row, &previousProduct); err != nil && err != sql.ErrNoRows {
		return model.Order{}, err
	}
	row = tx.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, input.ProductID)
	if err := scanProductInto(row, &nextProduct); err != nil {
		if err == sql.ErrNoRows {
			return model.Order{}, fmt.Errorf("未找到商品")
		}
		return model.Order{}, err
	}

	if previousProduct.ID != 0 && previousProduct.ID != nextProduct.ID {
		if _, err := tx.Exec(`UPDATE products SET stock = ? WHERE id = ?`,
			previousProduct.Stock+previousItem.Quantity, previousProduct.ID); err != nil {
			return model.Order{}, err
		}
	}

	available := nextProduct.Stock
	if nextProduct.ID == previousItem.ProductID {
		available += previousItem.Quantity
	}
	if available < input.Quantity {
		return model.Order{}, fmt.Errorf("库存不足，请先补货")
	}

	o := buildSingleOrder(nextProduct, input, existing.OrderNo, existing.CreatedAt, existing.Status)
	o.ID = id
	itemsJSON, err := marshal(o.Items)
	if err != nil {
		return model.Order{}, err
	}
	if _, err := tx.Exec(`UPDATE orders SET channel = ?, customer_name = ?, items_json = ?,
		total_amount = ?, discount_amount = ?, shipping_cost = ?, platform_fee = ?, goods_cost = ?,
		net_profit = ?, status = ? WHERE id = ?`,
		o.Channel, o.CustomerName, itemsJSON, o.TotalAmount, o.DiscountAmount, o.ShippingCost,
		o.PlatformFee, o.GoodsCost, o.NetProfit, o.Status, id); err != nil {
		return model.Order{}, err
	}
	if _, err := tx.Exec(`UPDATE products SET stock = ? WHERE id = ?`, available-input.Quantity, nextProduct.ID); err != nil {
		return model.Order{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Order{}, err
	}
	return o, nil
}

func (s *Store) DeleteOrder(id int64) error {
	existing, err := s.GetOrder(id)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(existing.Items) > 0 {
		item := existing.Items[0]
		if _, err := tx.Exec(`UPDATE products SET stock = stock + ? WHERE id = ?`, item.Quantity, item.ProductID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`DELETE FROM orders WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func scanProductInto(row rowScanner, p *model.Product) error {
	return row.Scan(&p.ID, &p.Name, &p.SKU, &p.Category, &p.Supplier, &p.PurchaseCost,
		&p.PackagingCost, &p.SalePrice, &p.Stock, &p.SafeStock, &p.Note, &p.CreatedAt)
}
