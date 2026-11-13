package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"onestarlab/backend/internal/model"
)

// dexieExportFile 是旧版前端通过 dexie-export-import 导出的 JSON 结构。
// 结构说明：data.tables 是表元信息，data.data 按表给出数据行；行内可能带有
// "$types" 元数据（例如数组字段），解析时会被忽略。
type dexieExportFile struct {
	FormatName    string `json:"formatName"`
	FormatVersion int    `json:"formatVersion"`
	Data          struct {
		DatabaseName    string `json:"databaseName"`
		DatabaseVersion int    `json:"databaseVersion"`
		Tables          []struct {
			Name string `json:"name"`
		} `json:"tables"`
		Data []struct {
			TableName string            `json:"tableName"`
			Inbound   bool              `json:"inbound"`
			Rows      []json.RawMessage `json:"rows"`
		} `json:"data"`
	} `json:"data"`
}

// ParseLegacyExport 把旧版浏览器（Dexie/IndexedDB）的导出文件转换成新版快照，
// 之后可以复用 ImportSnapshot 的事务导入逻辑。
func ParseLegacyExport(raw []byte) (model.Snapshot, error) {
	var snap model.Snapshot
	var file dexieExportFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return snap, fmt.Errorf("旧版备份解析失败：%v", err)
	}
	if file.FormatName != "dexie" {
		return snap, fmt.Errorf("不是旧版浏览器导出文件")
	}
	if file.FormatVersion != 1 {
		return snap, fmt.Errorf("不支持的旧版备份格式版本：%d", file.FormatVersion)
	}

	snap.Version = 1
	snap.ExportedAt = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")

	recognized := 0
	for _, table := range file.Data.Data {
		var err error
		switch strings.ToLower(strings.TrimSpace(table.TableName)) {
		case "products":
			snap.Products, err = decodeLegacyRows[model.Product](table.Rows)
		case "categories":
			snap.Categories, err = decodeLegacyRows[model.Category](table.Rows)
		case "orders":
			snap.Orders, err = decodeLegacyRows[model.Order](table.Rows)
		case "pricingrules":
			var rules []model.PricingRule
			rules, err = decodeLegacyRows[model.PricingRule](table.Rows)
			if err == nil && len(rules) > 0 {
				snap.PricingRule = rules[0]
			}
		case "labinspirations":
			snap.LabInspirations, err = decodeLegacyRows[model.LabInspiration](table.Rows)
		case "labprojects":
			snap.LabProjects, err = decodeLegacyRows[model.LabProject](table.Rows)
		case "drawinginspirations":
			snap.DrawingInspirations, err = decodeLegacyRows[model.DrawingInspirationRecord](table.Rows)
		case "expenses":
			snap.Expenses, err = decodeLegacyRows[model.ExpenseRecord](table.Rows)
		case "inventorytransactions":
			snap.InventoryTransactions, err = decodeLegacyRows[model.InventoryTransaction](table.Rows)
		case "markets":
			snap.Markets, err = decodeLegacyRows[model.MarketRecord](table.Rows)
		case "settings":
			snap.Settings, err = decodeLegacyRows[model.AppSetting](table.Rows)
		case "plannertemplates":
			snap.PlannerTemplates, err = decodeLegacyRows[model.PlannerTemplate](table.Rows)
		default:
			continue
		}
		if err != nil {
			return snap, fmt.Errorf("读取旧版数据表 %s 失败：%v", table.TableName, err)
		}
		recognized++
	}

	if recognized == 0 {
		return snap, fmt.Errorf("备份中没有可识别的 OneStarLab 数据表")
	}
	return snap, nil
}

// decodeLegacyRows 逐行解析旧版数据。Dexie 对 outbound 表会把行导出成
// [主键, 数据] 的二元数组，这里两种形态都兼容。
func decodeLegacyRows[T any](rows []json.RawMessage) ([]T, error) {
	out := make([]T, 0, len(rows))
	for index, raw := range rows {
		normalized, err := normalizeLegacyRow(raw)
		if err != nil {
			return nil, fmt.Errorf("第 %d 行：%v", index+1, err)
		}
		var item T
		if err := json.Unmarshal(normalized, &item); err != nil {
			return nil, fmt.Errorf("第 %d 行：%v", index+1, err)
		}
		out = append(out, item)
	}
	return out, nil
}

func normalizeLegacyRow(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("数据为空")
	}
	if trimmed[0] != '[' {
		return trimmed, nil
	}
	var pair []json.RawMessage
	if err := json.Unmarshal(trimmed, &pair); err != nil {
		return nil, err
	}
	if len(pair) < 2 {
		return nil, fmt.Errorf("无法识别的数据行")
	}
	return bytes.TrimSpace(pair[1]), nil
}
