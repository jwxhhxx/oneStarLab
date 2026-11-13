package store

import (
	"database/sql"
	"fmt"
	"strconv"

	"onestarlab/backend/internal/model"
)

func (s *Store) ListExpenses() ([]model.ExpenseRecord, error) {
	rows, err := s.db.Query(`SELECT id, purpose, amount, created_at FROM expenses ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.ExpenseRecord, 0)
	for rows.Next() {
		var e model.ExpenseRecord
		if err := rows.Scan(&e.ID, &e.Purpose, &e.Amount, &e.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, rows.Err()
}

func (s *Store) CreateExpense(input model.ExpenseInput) (model.ExpenseRecord, error) {
	createdAt := nowISO()
	res, err := s.db.Exec(`INSERT INTO expenses (purpose, amount, created_at) VALUES (?, ?, ?)`,
		input.Purpose, input.Amount, createdAt)
	if err != nil {
		return model.ExpenseRecord{}, err
	}
	id, _ := res.LastInsertId()
	return model.ExpenseRecord{ID: id, Purpose: input.Purpose, Amount: input.Amount, CreatedAt: createdAt}, nil
}

func (s *Store) UpdateExpense(id int64, input model.ExpenseInput) (model.ExpenseRecord, error) {
	res, err := s.db.Exec(`UPDATE expenses SET purpose = ?, amount = ? WHERE id = ?`, input.Purpose, input.Amount, id)
	if err != nil {
		return model.ExpenseRecord{}, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return model.ExpenseRecord{}, ErrNotFound
	}
	var e model.ExpenseRecord
	if err := s.db.QueryRow(`SELECT id, purpose, amount, created_at FROM expenses WHERE id = ?`, id).
		Scan(&e.ID, &e.Purpose, &e.Amount, &e.CreatedAt); err != nil {
		return model.ExpenseRecord{}, err
	}
	return e, nil
}

func (s *Store) DeleteExpense(id int64) error {
	res, err := s.db.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListMarkets() ([]model.MarketRecord, error) {
	rows, err := s.db.Query(`SELECT id, name, date, location, income, note, created_at
		FROM markets ORDER BY date DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.MarketRecord, 0)
	for rows.Next() {
		var m model.MarketRecord
		if err := rows.Scan(&m.ID, &m.Name, &m.Date, &m.Location, &m.Income, &m.Note, &m.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func (s *Store) CreateMarket(input model.MarketInput) (model.MarketRecord, error) {
	createdAt := nowISO()
	res, err := s.db.Exec(`INSERT INTO markets (name, date, location, income, note, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, input.Name, input.Date, input.Location, input.Income, input.Note, createdAt)
	if err != nil {
		return model.MarketRecord{}, err
	}
	id, _ := res.LastInsertId()
	return model.MarketRecord{ID: id, Name: input.Name, Date: input.Date, Location: input.Location,
		Income: input.Income, Note: input.Note, CreatedAt: createdAt}, nil
}

func (s *Store) UpdateMarket(id int64, input model.MarketInput) (model.MarketRecord, error) {
	res, err := s.db.Exec(`UPDATE markets SET name = ?, date = ?, location = ?, income = ?, note = ? WHERE id = ?`,
		input.Name, input.Date, input.Location, input.Income, input.Note, id)
	if err != nil {
		return model.MarketRecord{}, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return model.MarketRecord{}, ErrNotFound
	}
	return s.GetMarket(id)
}

func (s *Store) GetMarket(id int64) (model.MarketRecord, error) {
	var m model.MarketRecord
	err := s.db.QueryRow(`SELECT id, name, date, location, income, note, created_at FROM markets WHERE id = ?`, id).
		Scan(&m.ID, &m.Name, &m.Date, &m.Location, &m.Income, &m.Note, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return model.MarketRecord{}, ErrNotFound
	}
	return m, err
}

func (s *Store) DeleteMarket(id int64) error {
	res, err := s.db.Exec(`DELETE FROM markets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetOnlineIncome() (float64, string, error) {
	var value, updatedAt string
	err := s.db.QueryRow(`SELECT value, updated_at FROM settings WHERE key = 'onlineIncome'`).Scan(&value, &updatedAt)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		parsed = 0
	}
	return parsed, updatedAt, nil
}

func (s *Store) SaveOnlineIncome(amount float64) error {
	if amount < 0 {
		return fmt.Errorf("线上收入需为不小于 0 的数字")
	}
	_, err := s.db.Exec(`INSERT INTO settings (key, value, updated_at) VALUES ('onlineIncome', ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		strconv.FormatFloat(amount, 'f', -1, 64), nowISO())
	return err
}

func (s *Store) ListSettings() ([]model.AppSetting, error) {
	rows, err := s.db.Query(`SELECT key, value, updated_at FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.AppSetting, 0)
	for rows.Next() {
		var st model.AppSetting
		if err := rows.Scan(&st.Key, &st.Value, &st.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, rows.Err()
}

func (s *Store) ListPlannerTemplates() ([]model.PlannerTemplate, error) {
	rows, err := s.db.Query(`SELECT id, name, size_key, width_mm, height_mm, padding_mm, accent, paper,
		modules_json, created_at, updated_at FROM planner_templates ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]model.PlannerTemplate, 0)
	for rows.Next() {
		t, err := scanPlannerTemplate(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func scanPlannerTemplate(row rowScanner) (model.PlannerTemplate, error) {
	var t model.PlannerTemplate
	var modulesJSON string
	err := row.Scan(&t.ID, &t.Name, &t.SizeKey, &t.WidthMm, &t.HeightMm, &t.PaddingMm, &t.Accent,
		&t.Paper, &modulesJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return model.PlannerTemplate{}, err
	}
	if err := unmarshalItems(modulesJSON, &t.Modules); err != nil {
		return model.PlannerTemplate{}, err
	}
	if t.Modules == nil {
		t.Modules = []model.PlannerModuleConfig{}
	}
	return t, nil
}

func (s *Store) GetPlannerTemplate(id int64) (model.PlannerTemplate, error) {
	row := s.db.QueryRow(`SELECT id, name, size_key, width_mm, height_mm, padding_mm, accent, paper,
		modules_json, created_at, updated_at FROM planner_templates WHERE id = ?`, id)
	t, err := scanPlannerTemplate(row)
	if err == sql.ErrNoRows {
		return model.PlannerTemplate{}, ErrNotFound
	}
	return t, err
}

func (s *Store) SavePlannerTemplate(input model.PlannerTemplateInput, id int64) (model.PlannerTemplate, error) {
	modulesJSON, err := marshal(input.Modules)
	if err != nil {
		return model.PlannerTemplate{}, err
	}
	if id != 0 {
		_, err := s.db.Exec(`UPDATE planner_templates SET name = ?, size_key = ?, width_mm = ?, height_mm = ?,
			padding_mm = ?, accent = ?, paper = ?, modules_json = ?, updated_at = ? WHERE id = ?`,
			input.Name, input.SizeKey, input.WidthMm, input.HeightMm, input.PaddingMm, input.Accent,
			input.Paper, modulesJSON, nowISO(), id)
		if err != nil {
			return model.PlannerTemplate{}, err
		}
		return s.GetPlannerTemplate(id)
	}
	now := nowISO()
	res, err := s.db.Exec(`INSERT INTO planner_templates (name, size_key, width_mm, height_mm, padding_mm,
		accent, paper, modules_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.SizeKey, input.WidthMm, input.HeightMm, input.PaddingMm, input.Accent,
		input.Paper, modulesJSON, now, now)
	if err != nil {
		return model.PlannerTemplate{}, err
	}
	createdID, _ := res.LastInsertId()
	return s.GetPlannerTemplate(createdID)
}

func (s *Store) DeletePlannerTemplate(id int64) error {
	res, err := s.db.Exec(`DELETE FROM planner_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) Bootstrap() (model.Bootstrap, error) {
	var b model.Bootstrap
	var err error
	if b.Products, err = s.ListProducts(); err != nil {
		return b, err
	}
	if b.Categories, err = s.ListCategories(); err != nil {
		return b, err
	}
	if b.Orders, err = s.ListOrders(); err != nil {
		return b, err
	}
	if b.PricingRule, err = s.getPricingRule(); err != nil {
		return b, err
	}
	if b.LabInspirations, err = s.ListInspirations(); err != nil {
		return b, err
	}
	if b.LabProjects, err = s.ListLabProjects(); err != nil {
		return b, err
	}
	if b.DrawingInspirations, err = s.ListDrawingInspirations(); err != nil {
		return b, err
	}
	if b.Expenses, err = s.ListExpenses(); err != nil {
		return b, err
	}
	if b.InventoryTransactions, err = s.ListInventoryTransactions(); err != nil {
		return b, err
	}
	if b.Markets, err = s.ListMarkets(); err != nil {
		return b, err
	}
	if b.OnlineIncome, b.OnlineIncomeUpdatedAt, err = s.GetOnlineIncome(); err != nil {
		return b, err
	}
	if b.PlannerTemplates, err = s.ListPlannerTemplates(); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Store) Snapshot(exportedAt string) (model.Snapshot, error) {
	var snap model.Snapshot
	var err error
	snap.Version = 1
	snap.ExportedAt = exportedAt
	if snap.Products, err = s.ListProducts(); err != nil {
		return snap, err
	}
	if snap.Categories, err = s.ListCategories(); err != nil {
		return snap, err
	}
	if snap.Orders, err = s.ListOrders(); err != nil {
		return snap, err
	}
	if snap.PricingRule, err = s.getPricingRule(); err != nil {
		return snap, err
	}
	if snap.LabInspirations, err = s.ListInspirations(); err != nil {
		return snap, err
	}
	if snap.LabProjects, err = s.ListLabProjects(); err != nil {
		return snap, err
	}
	if snap.DrawingInspirations, err = s.ListDrawingInspirations(); err != nil {
		return snap, err
	}
	if snap.Expenses, err = s.ListExpenses(); err != nil {
		return snap, err
	}
	if snap.InventoryTransactions, err = s.ListInventoryTransactions(); err != nil {
		return snap, err
	}
	if snap.Markets, err = s.ListMarkets(); err != nil {
		return snap, err
	}
	if snap.Settings, err = s.ListSettings(); err != nil {
		return snap, err
	}
	if snap.PlannerTemplates, err = s.ListPlannerTemplates(); err != nil {
		return snap, err
	}
	return snap, nil
}

// ImportSnapshot 用快照整体替换数据库内容，全过程在一个事务内完成。
func (s *Store) ImportSnapshot(snap model.Snapshot) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tables := []string{
		"products", "categories", "orders", "pricing_rules", "lab_inspirations", "lab_projects",
		"drawing_inspirations", "expenses", "inventory_transactions", "markets", "settings",
		"planner_templates",
	}
	for _, table := range tables {
		if _, err := tx.Exec(`DELETE FROM ` + table); err != nil {
			return err
		}
	}

	for _, p := range snap.Products {
		if _, err := tx.Exec(`INSERT INTO products (id, name, sku, category, supplier, purchase_cost,
			packaging_cost, sale_price, stock, safe_stock, note, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ID, p.Name, p.SKU, p.Category, p.Supplier, p.PurchaseCost, p.PackagingCost, p.SalePrice,
			p.Stock, p.SafeStock, p.Note, p.CreatedAt); err != nil {
			return err
		}
	}
	for _, c := range snap.Categories {
		if _, err := tx.Exec(`INSERT INTO categories (id, name, created_at) VALUES (?, ?, ?)`,
			c.ID, c.Name, c.CreatedAt); err != nil {
			return err
		}
	}
	for _, o := range snap.Orders {
		itemsJSON, err := marshal(o.Items)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO orders (id, order_no, channel, customer_name, items_json,
			total_amount, discount_amount, shipping_cost, platform_fee, goods_cost, net_profit, created_at, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, o.ID, o.OrderNo, o.Channel, o.CustomerName,
			itemsJSON, o.TotalAmount, o.DiscountAmount, o.ShippingCost, o.PlatformFee, o.GoodsCost,
			o.NetProfit, o.CreatedAt, o.Status); err != nil {
			return err
		}
	}
	if snap.PricingRule.ID == 0 {
		snap.PricingRule = model.PricingRule{ID: 1, TargetMargin: 0.45, ChannelFeeRate: 0.05, ExtraCost: 1.2, RoundingMode: "x.9"}
	}
	if _, err := tx.Exec(`INSERT INTO pricing_rules (id, target_margin, channel_fee_rate, extra_cost, rounding_mode)
		VALUES (1, ?, ?, ?, ?)`, snap.PricingRule.TargetMargin, snap.PricingRule.ChannelFeeRate,
		snap.PricingRule.ExtraCost, snap.PricingRule.RoundingMode); err != nil {
		return err
	}
	for _, it := range snap.LabInspirations {
		if _, err := tx.Exec(`INSERT INTO lab_inspirations (id, title, tag, status, summary, keywords, image,
			created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, it.ID, it.Title, it.Tag, it.Status,
			it.Summary, it.Keywords, it.Image, it.CreatedAt, it.UpdatedAt); err != nil {
			return err
		}
	}
	for _, p := range snap.LabProjects {
		samples, err := marshal(p.SampleRecords)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO lab_projects (id, name, category, stage, progress, launch_date, note,
			image, sample_records_json, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.ID, p.Name,
			p.Category, p.Stage, p.Progress, p.LaunchDate, p.Note, p.Image, samples, p.UpdatedAt); err != nil {
			return err
		}
	}
	for _, r := range snap.DrawingInspirations {
		items, err := marshal(r.Items)
		if err != nil {
			return err
		}
		styles, err := marshal(r.Styles)
		if err != nil {
			return err
		}
		completed := 0
		if r.Completed {
			completed = 1
		}
		if _, err := tx.Exec(`INSERT INTO drawing_inspirations (id, items_json, styles_json, created_at, completed)
			VALUES (?, ?, ?, ?, ?)`, r.ID, items, styles, r.CreatedAt, completed); err != nil {
			return err
		}
	}
	for _, e := range snap.Expenses {
		if _, err := tx.Exec(`INSERT INTO expenses (id, purpose, amount, created_at) VALUES (?, ?, ?, ?)`,
			e.ID, e.Purpose, e.Amount, e.CreatedAt); err != nil {
			return err
		}
	}
	for _, t := range snap.InventoryTransactions {
		if _, err := tx.Exec(`INSERT INTO inventory_transactions (id, product_id, barcode, type, quantity, note,
			created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, t.ID, t.ProductID, t.Barcode, t.Type, t.Quantity,
			t.Note, t.CreatedAt); err != nil {
			return err
		}
	}
	for _, m := range snap.Markets {
		if _, err := tx.Exec(`INSERT INTO markets (id, name, date, location, income, note, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, m.ID, m.Name, m.Date, m.Location, m.Income, m.Note, m.CreatedAt); err != nil {
			return err
		}
	}
	for _, st := range snap.Settings {
		if _, err := tx.Exec(`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)`,
			st.Key, st.Value, st.UpdatedAt); err != nil {
			return err
		}
	}
	for _, t := range snap.PlannerTemplates {
		modules, err := marshal(t.Modules)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO planner_templates (id, name, size_key, width_mm, height_mm, padding_mm,
			accent, paper, modules_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			t.ID, t.Name, t.SizeKey, t.WidthMm, t.HeightMm, t.PaddingMm, t.Accent, t.Paper, modules,
			t.CreatedAt, t.UpdatedAt); err != nil {
			return err
		}
	}

	// 让自增序列重新跟随当前最大 id，避免后续插入冲突。
	if _, err := tx.Exec(`DELETE FROM sqlite_sequence WHERE name IN
		('products','categories','orders','lab_inspirations','lab_projects','drawing_inspirations',
		 'expenses','inventory_transactions','markets','planner_templates')`); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	// 恢复必备的默认分类（快照里没有时）。
	if _, err := s.ensureCategory("未分类"); err != nil {
		return err
	}
	return nil
}
