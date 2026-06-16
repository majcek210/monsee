package service

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/encrypt"
	"github.com/majcek210/monsee/pkg/hash"
)

type TwoFactorService struct {
	users  domain.UserRepository
	encKey []byte
}

func NewTwoFactorService(users domain.UserRepository, encKey []byte) *TwoFactorService {
	return &TwoFactorService{users: users, encKey: encKey}
}

func (s *TwoFactorService) InitiateSetup(ctx context.Context, userID string) (secret, otpauthURI string, err error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "monsee",
		AccountName: user.Email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generate totp: %w", err)
	}

	encrypted, err := encrypt.Encrypt(s.encKey, key.Secret())
	if err != nil {
		return "", "", fmt.Errorf("encrypt totp secret: %w", err)
	}

	if err := s.users.SetTOTPSecret(ctx, userID, encrypted); err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

func (s *TwoFactorService) ConfirmSetup(ctx context.Context, userID, code string) ([]string, error) {
	totpData, err := s.users.GetTOTP(ctx, userID)
	if err != nil {
		return nil, err
	}
	if totpData.Secret == nil {
		return nil, domain.ValidationErr("code", "2FA setup not initiated")
	}

	secret, err := encrypt.Decrypt(s.encKey, *totpData.Secret)
	if err != nil {
		return nil, fmt.Errorf("decrypt totp secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return nil, domain.ValidationErr("code", "invalid 2FA code")
	}

	backupCodes := make([]string, 10)
	hashedCodes := make([]string, 10)
	for i := range backupCodes {
		raw, err := generateBackupCode()
		if err != nil {
			return nil, err
		}
		backupCodes[i] = raw
		hashedCodes[i] = hash.SHA256(raw)
	}

	if err := s.users.EnableTOTP(ctx, userID, hashedCodes); err != nil {
		return nil, err
	}

	return backupCodes, nil
}

func (s *TwoFactorService) Disable(ctx context.Context, userID, password string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if !hash.CheckPassword(user.PasswordHash, password) {
		return domain.Unauthorized("invalid password")
	}
	return s.users.DisableTOTP(ctx, userID)
}

func (s *TwoFactorService) Verify(ctx context.Context, userID, code string) (bool, error) {
	totpData, err := s.users.GetTOTP(ctx, userID)
	if err != nil {
		return false, err
	}
	if !totpData.Enabled || totpData.Secret == nil {
		return false, domain.ValidationErr("code", "2FA not enabled")
	}

	secret, err := encrypt.Decrypt(s.encKey, *totpData.Secret)
	if err != nil {
		return false, fmt.Errorf("decrypt totp secret: %w", err)
	}

	if totp.Validate(code, secret) {
		return true, nil
	}

	codeHash := hash.SHA256(code)
	for _, bc := range totpData.BackupCodes {
		if bc == codeHash {
			if err := s.users.RemoveBackupCode(ctx, userID, bc); err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, nil
}

func generateBackupCode() (string, error) {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

var _ = time.Now
