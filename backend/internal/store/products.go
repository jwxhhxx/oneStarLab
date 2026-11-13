package store

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"onestarlab/backend/internal/model"
)

const maxCategoryNameLength = 10

func (s *Store) ListProducts() ([]model.Product, error) {
	rows, err := s.db.Query(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (model.Product, error) {
	var p model.Product
	err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.Category, &p.Supplier, &p.PurchaseCost,
		&p.PackagingCost, &p.SalePrice, &p.Stock, &p.SafeStock, &p.Note, &p.CreatedAt)
	return p, err
}

func (s *Store) GetProduct(id int64) (model.Product, error) {
	row := s.db.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE id = ?`, id)
	p, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return model.Product{}, ErrNotFound
	}
	return p, err
}

func (s *Store) FindProductBySKU(sku string) (model.Product, error) {
	row := s.db.QueryRow(`SELECT id, name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at FROM products WHERE sku = ?`, sku)
	p, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return model.Product{}, ErrNotFound
	}
	return p, err
}

func (s *Store) CreateProduct(input model.ProductInput) (model.Product, error) {
	rule, err := s.getPricingRule()
	if err != nil {
		return model.Product{}, err
	}

	category := strings.TrimSpace(input.Category)
	if category != "" {
		if _, err := s.ensureCategory(category); err != nil {
			return model.Product{}, err
		}
	}

	sku := strings.TrimSpace(input.SKU)
	if sku == "" {
		sku, err = s.generateSKU(category, 0, true)
		if err != nil {
			return model.Product{}, err
		}
	}

	salePrice := input.SalePrice
	if salePrice <= 0 {
		salePrice = calculateSuggestedPrice(input.PurchaseCost, input.PackagingCost, rule)
	}

	createdAt := nowISO()
	res, err := s.db.Exec(`INSERT INTO products
		(name, sku, category, supplier, purchase_cost, packaging_cost, sale_price, stock, safe_stock, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		strings.TrimSpace(input.Name), sku, category, input.Supplier, input.PurchaseCost,
		input.PackagingCost, salePrice, input.Stock, input.SafeStock, input.Note, createdAt,
	)
	if err != nil {
		return model.Product{}, fmt.Errorf("新增商品失败: %w", err)
	}
	id, _ := res.LastInsertId()
	return s.GetProduct(id)
}

func (s *Store) UpdateProduct(id int64, input model.ProductInput) (model.Product, error) {
	rule, err := s.getPricingRule()
	if err != nil {
		return model.Product{}, err
	}
	category := strings.TrimSpace(input.Category)
	if category != "" {
		if _, err := s.ensureCategory(category); err != nil {
			return model.Product{}, err
		}
	}
	salePrice := input.SalePrice
	if salePrice <= 0 {
		salePrice = calculateSuggestedPrice(input.PurchaseCost, input.PackagingCost, rule)
	}

	res, err := s.db.Exec(`UPDATE products SET name = ?, sku = ?, category = ?, supplier = ?,
		purchase_cost = ?, packaging_cost = ?, sale_price = ?, stock = ?, safe_stock = ?, note = ?
		WHERE id = ?`,
		strings.TrimSpace(input.Name), strings.TrimSpace(input.SKU), category, input.Supplier,
		input.PurchaseCost, input.PackagingCost, salePrice, input.Stock, input.SafeStock, input.Note, id,
	)
	if err != nil {
		return model.Product{}, fmt.Errorf("更新商品失败: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return model.Product{}, ErrNotFound
	}
	return s.GetProduct(id)
}

func (s *Store) DeleteProduct(id int64) error {
	res, err := s.db.Exec(`DELETE FROM products WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListCategories() ([]model.Category, error) {
	rows, err := s.db.Query(`SELECT id, name, created_at FROM categories ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]model.Category, 0)
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func clampCategoryName(name string) string {
	trimmed := strings.TrimSpace(name)
	runes := []rune(trimmed)
	if len(runes) > maxCategoryNameLength {
		return string(runes[:maxCategoryNameLength])
	}
	return trimmed
}

// ensureCategory 保证分类存在，返回其记录。
func (s *Store) ensureCategory(name string) (model.Category, error) {
	nm := clampCategoryName(name)
	if nm == "" {
		return model.Category{}, nil
	}
	var c model.Category
	err := s.db.QueryRow(`SELECT id, name, created_at FROM categories WHERE name = ?`, nm).
		Scan(&c.ID, &c.Name, &c.CreatedAt)
	if err == nil {
		return c, nil
	}
	if err != sql.ErrNoRows {
		return model.Category{}, err
	}
	res, err := s.db.Exec(`INSERT OR IGNORE INTO categories (name, created_at) VALUES (?, ?)`, nm, nowISO())
	if err != nil {
		return model.Category{}, err
	}
	id, _ := res.LastInsertId()
	if id == 0 {
		if err := s.db.QueryRow(`SELECT id FROM categories WHERE name = ?`, nm).Scan(&id); err != nil {
			return model.Category{}, err
		}
	}
	return model.Category{ID: id, Name: nm, CreatedAt: nowISO()}, nil
}

func (s *Store) AddCategory(name string) (model.Category, error) {
	return s.ensureCategory(name)
}

// DeleteCategory 删除分类；reassignTo 非空时把商品迁移过去，否则清空商品分类。
func (s *Store) DeleteCategory(id int64, reassignTo string) error {
	var cat model.Category
	err := s.db.QueryRow(`SELECT id, name, created_at FROM categories WHERE id = ?`, id).
		Scan(&cat.ID, &cat.Name, &cat.CreatedAt)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if cat.ID == 0 || cat.Name == "未分类" {
		return fmt.Errorf("默认分类“未分类”不可删除")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM products WHERE category = ?`, cat.Name).Scan(&count); err != nil {
		return err
	}

	target := clampCategoryName(reassignTo)
	if count > 0 {
		if target != "" {
			var exists int
			err := tx.QueryRow(`SELECT id FROM categories WHERE name = ?`, target).Scan(&exists)
			if err == sql.ErrNoRows {
				if _, err := tx.Exec(`INSERT INTO categories (name, created_at) VALUES (?, ?)`, target, nowISO()); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE products SET category = ? WHERE category = ?`, target, cat.Name); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(`UPDATE products SET category = '' WHERE category = ?`, cat.Name); err != nil {
				return err
			}
		}
	}

	if _, err := tx.Exec(`DELETE FROM categories WHERE id = ?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// GenerateSKU 按「分类ID-四位序号」生成最小空缺序号。
func (s *Store) GenerateSKU(category string, excludeProductID int64, createIfMissing bool) (string, error) {
	return s.generateSKU(category, excludeProductID, createIfMissing)
}

func (s *Store) generateSKU(category string, excludeProductID int64, createIfMissing bool) (string, error) {
	nm := strings.TrimSpace(category)
	var categoryID int64
	if nm != "" {
		var id int64
		err := s.db.QueryRow(`SELECT id FROM categories WHERE name = ?`, nm).Scan(&id)
		if err == sql.ErrNoRows {
			if createIfMissing {
				cat, err := s.ensureCategory(nm)
				if err != nil {
					return "", err
				}
				categoryID = cat.ID
			}
		} else if err != nil {
			return "", err
		} else {
			categoryID = id
		}
	}

	query := `SELECT id, sku FROM products`
	args := []any{}
	if nm != "" {
		query += ` WHERE category = ?`
		args = append(args, nm)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	pattern := regexp.MustCompile(fmt.Sprintf(`^%d-(\d+)$`, categoryID))
	used := map[int64]bool{}
	for rows.Next() {
		var id int64
		var sku string
		if err := rows.Scan(&id, &sku); err != nil {
			return "", err
		}
		if excludeProductID != 0 && id == excludeProductID {
			continue
		}
		m := pattern.FindStringSubmatch(sku)
		if len(m) == 2 {
			if n, err := strconv.ParseInt(m[1], 10, 64); err == nil && n > 0 {
				used[n] = true
			}
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	seq := int64(1)
	for used[seq] {
		seq++
	}
	candidate := fmt.Sprintf("%d-%04d", categoryID, seq)

	for {
		var existsID int64
		err := s.db.QueryRow(`SELECT id FROM products WHERE sku = ?`, candidate).Scan(&existsID)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return "", err
		}
		if excludeProductID != 0 && existsID == excludeProductID {
			break
		}
		seq++
		candidate = fmt.Sprintf("%d-%04d", categoryID, seq)
	}
	return candidate, nil
}
