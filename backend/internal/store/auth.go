package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"onestarlab/backend/internal/auth"
)

// ErrInvalidPasscode 表示当前口令不正确。
var ErrInvalidPasscode = errors.New("口令不正确")

const (
	settingPasscodeSalt       = "auth.passcodeSalt"
	settingPasscodeHash       = "auth.passcodeHash"
	settingPasscodeIterations = "auth.passcodeIterations"
	settingPasscodeIsDefault  = "auth.passcodeIsDefault"
	settingSessionSecret      = "auth.sessionSecret"
)

func (s *Store) getSetting(key string) (string, bool, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (s *Store) putSetting(tx *sql.Tx, key, value string) error {
	_, err := tx.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, nowISO(),
	)
	return err
}

// VerifyPasscode 校验登录口令。
func (s *Store) VerifyPasscode(passcode string) (bool, error) {
	salt, ok, err := s.getSetting(settingPasscodeSalt)
	if err != nil || !ok {
		return false, err
	}
	hash, ok, err := s.getSetting(settingPasscodeHash)
	if err != nil || !ok {
		return false, err
	}
	iterations := auth.PasscodeIterations
	if raw, ok, err := s.getSetting(settingPasscodeIterations); err != nil {
		return false, err
	} else if ok {
		if parsed, parseErr := strconv.Atoi(raw); parseErr == nil && parsed > 0 {
			iterations = parsed
		}
	}
	return auth.VerifyPasscode(strings.TrimSpace(passcode), auth.Credentials{
		Salt:       salt,
		Hash:       hash,
		Iterations: iterations,
	}), nil
}

// PasscodeIsDefault 表示当前是否仍在使用初始口令。
func (s *Store) PasscodeIsDefault() bool {
	value, ok, err := s.getSetting(settingPasscodeIsDefault)
	return err == nil && ok && value == "1"
}

// ChangePasscode 校验当前口令后写入新口令，并轮换会话密钥。
func (s *Store) ChangePasscode(current, next string) error {
	ok, err := s.VerifyPasscode(current)
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidPasscode
	}
	if err := auth.ValidatePasscode(next); err != nil {
		return err
	}
	if current == next {
		return errors.New("新口令不能与当前口令相同")
	}

	credentials, err := auth.HashPasscode(next)
	if err != nil {
		return err
	}
	secret, err := auth.RandomToken(32)
	if err != nil {
		return fmt.Errorf("生成会话密钥失败: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	values := map[string]string{
		settingPasscodeSalt:       credentials.Salt,
		settingPasscodeHash:       credentials.Hash,
		settingPasscodeIterations: strconv.Itoa(credentials.Iterations),
		settingPasscodeIsDefault:  "0",
		settingSessionSecret:      secret,
	}
	for key, value := range values {
		if err := s.putSetting(tx, key, value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// SessionSecret 返回用于签名会话令牌的密钥。
func (s *Store) SessionSecret() (string, error) {
	value, ok, err := s.getSetting(settingSessionSecret)
	if err != nil {
		return "", err
	}
	if !ok || value == "" {
		return "", errors.New("会话密钥未初始化")
	}
	return value, nil
}
