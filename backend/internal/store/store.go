package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"onestarlab/backend/internal/model"
)

// Store 封装所有业务数据访问，保证写操作在事务中完成。
type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// ErrNotFound 表示目标记录不存在。
var ErrNotFound = errors.New("记录不存在")

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundPrice(value float64, mode string) float64 {
	switch mode {
	case "whole":
		return math.Ceil(value)
	case "x.9":
		candidate := math.Floor(value) + 0.9
		if candidate >= value {
			return candidate
		}
		return math.Ceil(value) + 0.9
	default:
		return value
	}
}

func calculateSuggestedPrice(purchaseCost, packagingCost float64, rule model.PricingRule) float64 {
	totalCost := purchaseCost + packagingCost + rule.ExtraCost
	denominator := 1 - rule.TargetMargin - rule.ChannelFeeRate
	raw := totalCost
	if denominator > 0 {
		raw = totalCost / denominator
	}
	return round2(roundPrice(raw, rule.RoundingMode))
}

func marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalItems(raw string, target any) error {
	if strings.TrimSpace(raw) == "" {
		raw = "[]"
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("解析 JSON 字段失败: %w", err)
	}
	return nil
}

func (s *Store) getPricingRule() (model.PricingRule, error) {
	var rule model.PricingRule
	err := s.db.QueryRow(
		`SELECT id, target_margin, channel_fee_rate, extra_cost, rounding_mode FROM pricing_rules WHERE id = 1`,
	).Scan(&rule.ID, &rule.TargetMargin, &rule.ChannelFeeRate, &rule.ExtraCost, &rule.RoundingMode)
	if err == sql.ErrNoRows {
		rule = model.PricingRule{ID: 1, TargetMargin: 0.45, ChannelFeeRate: 0.05, ExtraCost: 1.2, RoundingMode: "x.9"}
		return rule, nil
	}
	return rule, err
}

// GetPricingRule 返回当前定价规则（不存在时返回默认值）。
func (s *Store) GetPricingRule() (model.PricingRule, error) {
	return s.getPricingRule()
}

func (s *Store) SavePricingRule(rule model.PricingRule) error {
	rule.ID = 1
	_, err := s.db.Exec(
		`INSERT INTO pricing_rules (id, target_margin, channel_fee_rate, extra_cost, rounding_mode)
		 VALUES (1, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET target_margin = excluded.target_margin,
		   channel_fee_rate = excluded.channel_fee_rate, extra_cost = excluded.extra_cost,
		   rounding_mode = excluded.rounding_mode`,
		rule.TargetMargin, rule.ChannelFeeRate, rule.ExtraCost, rule.RoundingMode,
	)
	return err
}
