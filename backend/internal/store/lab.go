package store

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"onestarlab/backend/internal/model"
)

var stageProgress = map[string]int{
	"灵感整理": 10,
	"设计中":  35,
	"打样中":  60,
	"排产中":  82,
	"待上新":  95,
	"已上新":  100,
}

func clampProgress(stage string, progress int) int {
	if min, ok := stageProgress[stage]; ok && progress < min {
		return min
	}
	return progress
}

func (s *Store) ListInspirations() ([]model.LabInspiration, error) {
	rows, err := s.db.Query(`SELECT id, title, tag, status, summary, keywords, image, created_at, updated_at
		FROM lab_inspirations ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.LabInspiration, 0)
	for rows.Next() {
		var it model.LabInspiration
		if err := rows.Scan(&it.ID, &it.Title, &it.Tag, &it.Status, &it.Summary, &it.Keywords,
			&it.Image, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, it)
	}
	return list, rows.Err()
}

func (s *Store) CreateInspiration(input model.InspirationInput) (model.LabInspiration, error) {
	now := nowISO()
	res, err := s.db.Exec(`INSERT INTO lab_inspirations (title, tag, status, summary, keywords, image, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, input.Title, input.Tag, input.Status, input.Summary,
		input.Keywords, input.Image, now, now)
	if err != nil {
		return model.LabInspiration{}, err
	}
	id, _ := res.LastInsertId()
	return model.LabInspiration{ID: id, Title: input.Title, Tag: input.Tag, Status: input.Status,
		Summary: input.Summary, Keywords: input.Keywords, Image: input.Image, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Store) UpdateInspiration(id int64, input model.InspirationInput) (model.LabInspiration, error) {
	res, err := s.db.Exec(`UPDATE lab_inspirations SET title = ?, tag = ?, status = ?, summary = ?,
		keywords = ?, image = ?, updated_at = ? WHERE id = ?`,
		input.Title, input.Tag, input.Status, input.Summary, input.Keywords, input.Image, nowISO(), id)
	if err != nil {
		return model.LabInspiration{}, err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return model.LabInspiration{}, ErrNotFound
	}
	return s.GetInspiration(id)
}

func (s *Store) GetInspiration(id int64) (model.LabInspiration, error) {
	var it model.LabInspiration
	err := s.db.QueryRow(`SELECT id, title, tag, status, summary, keywords, image, created_at, updated_at
		FROM lab_inspirations WHERE id = ?`, id).Scan(&it.ID, &it.Title, &it.Tag, &it.Status,
		&it.Summary, &it.Keywords, &it.Image, &it.CreatedAt, &it.UpdatedAt)
	if err == sql.ErrNoRows {
		return model.LabInspiration{}, ErrNotFound
	}
	return it, err
}

func (s *Store) DeleteInspiration(id int64) error {
	res, err := s.db.Exec(`DELETE FROM lab_inspirations WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListLabProjects() ([]model.LabProject, error) {
	rows, err := s.db.Query(`SELECT id, name, category, stage, progress, launch_date, note, image,
		sample_records_json, updated_at FROM lab_projects ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.LabProject, 0)
	for rows.Next() {
		p, err := scanLabProject(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func scanLabProject(row rowScanner) (model.LabProject, error) {
	var p model.LabProject
	var samplesJSON string
	err := row.Scan(&p.ID, &p.Name, &p.Category, &p.Stage, &p.Progress, &p.LaunchDate, &p.Note,
		&p.Image, &samplesJSON, &p.UpdatedAt)
	if err != nil {
		return model.LabProject{}, err
	}
	if err := unmarshalItems(samplesJSON, &p.SampleRecords); err != nil {
		return model.LabProject{}, err
	}
	if p.SampleRecords == nil {
		p.SampleRecords = []model.ProofRecord{}
	}
	return p, nil
}

func (s *Store) GetLabProject(id int64) (model.LabProject, error) {
	row := s.db.QueryRow(`SELECT id, name, category, stage, progress, launch_date, note, image,
		sample_records_json, updated_at FROM lab_projects WHERE id = ?`, id)
	p, err := scanLabProject(row)
	if err == sql.ErrNoRows {
		return model.LabProject{}, ErrNotFound
	}
	return p, err
}

func (s *Store) CreateLabProject(input model.LabProjectInput) (model.LabProject, error) {
	now := nowISO()
	samples, err := marshal([]model.ProofRecord{})
	if err != nil {
		return model.LabProject{}, err
	}
	res, err := s.db.Exec(`INSERT INTO lab_projects (name, category, stage, progress, launch_date, note, image,
		sample_records_json, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, input.Category, input.Stage, clampProgress(input.Stage, input.Progress),
		input.LaunchDate, input.Note, input.Image, samples, now)
	if err != nil {
		return model.LabProject{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetLabProject(id)
}

func (s *Store) UpdateLabProject(id int64, input model.LabProjectInput) (model.LabProject, error) {
	res, err := s.db.Exec(`UPDATE lab_projects SET name = ?, category = ?, stage = ?, progress = ?,
		launch_date = ?, note = ?, image = ?, updated_at = ? WHERE id = ?`,
		input.Name, input.Category, input.Stage, clampProgress(input.Stage, input.Progress),
		input.LaunchDate, input.Note, input.Image, nowISO(), id)
	if err != nil {
		return model.LabProject{}, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return model.LabProject{}, ErrNotFound
	}
	return s.GetLabProject(id)
}

func (s *Store) UpdateLabProjectStage(id int64, stage string) (model.LabProject, error) {
	existing, err := s.GetLabProject(id)
	if err != nil {
		return model.LabProject{}, err
	}
	progress := existing.Progress
	if min, ok := stageProgress[stage]; ok && progress < min {
		progress = min
	}
	res, err := s.db.Exec(`UPDATE lab_projects SET stage = ?, progress = ?, updated_at = ? WHERE id = ?`,
		stage, progress, nowISO(), id)
	if err != nil {
		return model.LabProject{}, err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return model.LabProject{}, ErrNotFound
	}
	return s.GetLabProject(id)
}

func (s *Store) AddProofRecord(projectID int64, input model.ProofRecordInput) (model.LabProject, error) {
	project, err := s.GetLabProject(projectID)
	if err != nil {
		return model.LabProject{}, err
	}
	record := model.ProofRecord{
		ID:    "proof-" + strconv.FormatInt(time.Now().UnixMilli(), 10),
		Title: input.Title, Date: input.Date, Note: input.Note, Image: input.Image,
	}
	project.SampleRecords = append(project.SampleRecords, record)
	samples, err := marshal(project.SampleRecords)
	if err != nil {
		return model.LabProject{}, err
	}
	if _, err := s.db.Exec(`UPDATE lab_projects SET sample_records_json = ?, updated_at = ? WHERE id = ?`,
		samples, nowISO(), projectID); err != nil {
		return model.LabProject{}, err
	}
	return s.GetLabProject(projectID)
}

func (s *Store) DeleteLabProject(id int64) error {
	res, err := s.db.Exec(`DELETE FROM lab_projects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ConvertLabProjectToProduct(id int64) (model.Product, error) {
	project, err := s.GetLabProject(id)
	if err != nil {
		return model.Product{}, err
	}
	sku := fmt.Sprintf("LAB-%d", id)
	var exists int64
	if err := s.db.QueryRow(`SELECT id FROM products WHERE sku = ?`, sku).Scan(&exists); err == nil {
		return model.Product{}, fmt.Errorf("该研发产品已转为正式商品")
	} else if err != sql.ErrNoRows {
		return model.Product{}, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return model.Product{}, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO products (name, sku, category, supplier, purchase_cost, packaging_cost,
		sale_price, stock, safe_stock, note, created_at) VALUES (?, ?, ?, ?, 0, 0, 0, 0, 5, ?, ?)`,
		project.Name, sku, project.Category, "待确认供应商", "由研究所转入："+project.Note, nowISO())
	if err != nil {
		return model.Product{}, err
	}
	if _, err := tx.Exec(`UPDATE lab_projects SET stage = '已上新', progress = 100, updated_at = ? WHERE id = ?`,
		nowISO(), id); err != nil {
		return model.Product{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Product{}, err
	}
	productID, _ := res.LastInsertId()
	return s.GetProduct(productID)
}

func (s *Store) ListDrawingInspirations() ([]model.DrawingInspirationRecord, error) {
	rows, err := s.db.Query(`SELECT id, items_json, styles_json, created_at, completed
		FROM drawing_inspirations ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]model.DrawingInspirationRecord, 0)
	for rows.Next() {
		var r model.DrawingInspirationRecord
		var itemsJSON, stylesJSON string
		var completed int
		if err := rows.Scan(&r.ID, &itemsJSON, &stylesJSON, &r.CreatedAt, &completed); err != nil {
			return nil, err
		}
		if err := unmarshalItems(itemsJSON, &r.Items); err != nil {
			return nil, err
		}
		if err := unmarshalItems(stylesJSON, &r.Styles); err != nil {
			return nil, err
		}
		r.Completed = completed == 1
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) CreateDrawingInspiration(input model.DrawingInspirationRecord) (model.DrawingInspirationRecord, error) {
	items, err := marshal(input.Items)
	if err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	styles, err := marshal(input.Styles)
	if err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	createdAt := nowISO()
	completed := 0
	if input.Completed {
		completed = 1
	}
	res, err := s.db.Exec(`INSERT INTO drawing_inspirations (items_json, styles_json, created_at, completed)
		VALUES (?, ?, ?, ?)`, items, styles, createdAt, completed)
	if err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	id, _ := res.LastInsertId()
	input.ID = id
	input.CreatedAt = createdAt
	return input, nil
}

func (s *Store) UpdateDrawingInspiration(id int64, patch model.DrawingInspirationRecord) error {
	existing, err := s.GetDrawingInspiration(id)
	if err != nil {
		return err
	}
	items := patch.Items
	if items == nil {
		items = existing.Items
	}
	styles := patch.Styles
	if styles == nil {
		styles = existing.Styles
	}
	itemsJSON, err := marshal(items)
	if err != nil {
		return err
	}
	stylesJSON, err := marshal(styles)
	if err != nil {
		return err
	}
	completed := existing.Completed
	if patch.Completed {
		completed = true
	}
	flag := 0
	if completed {
		flag = 1
	}
	res, err := s.db.Exec(`UPDATE drawing_inspirations SET items_json = ?, styles_json = ?, completed = ?
		WHERE id = ?`, itemsJSON, stylesJSON, flag, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) GetDrawingInspiration(id int64) (model.DrawingInspirationRecord, error) {
	var r model.DrawingInspirationRecord
	var itemsJSON, stylesJSON string
	var completed int
	err := s.db.QueryRow(`SELECT id, items_json, styles_json, created_at, completed
		FROM drawing_inspirations WHERE id = ?`, id).
		Scan(&r.ID, &itemsJSON, &stylesJSON, &r.CreatedAt, &completed)
	if err == sql.ErrNoRows {
		return model.DrawingInspirationRecord{}, ErrNotFound
	}
	if err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	if err := unmarshalItems(itemsJSON, &r.Items); err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	if err := unmarshalItems(stylesJSON, &r.Styles); err != nil {
		return model.DrawingInspirationRecord{}, err
	}
	r.Completed = completed == 1
	return r, nil
}

func (s *Store) DeleteDrawingInspiration(id int64) error {
	res, err := s.db.Exec(`DELETE FROM drawing_inspirations WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}
