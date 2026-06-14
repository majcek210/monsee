package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/majcek210/monsee/internal/domain"
	"github.com/majcek210/monsee/pkg/hash"
)

type APIKeyService struct {
	keys domain.APIKeyRepository
}

func NewAPIKeyService(keys domain.APIKeyRepository) *APIKeyService {
	return &APIKeyService{keys: keys}
}

// Generate creates a new API key, stores its SHA-256 hash, and returns the
// plain-text key (shown only once).
func (s *APIKeyService) Generate(ctx context.Context, userID, name string) (plainKey string, key *domain.APIKey, err error) {
	if name == "" {
		return "", nil, domain.ValidationErr("name", "name is required")
	}

	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("generate key bytes: %w", err)
	}

	plain := "sk_" + hex.EncodeToString(b)
	prefix := plain[:8] // "sk_" + first 5 hex chars
	keyHash := hash.SHA256(plain)

	k, err := s.keys.Create(ctx, domain.CreateAPIKeyParams{
		UserID:  userID,
		Name:    name,
		KeyHash: keyHash,
		Prefix:  prefix,
	})
	if err != nil {
		return "", nil, err
	}
	return plain, k, nil
}

// ValidateAndGetUserID hashes the raw key and looks it up.
// On success it updates last_used asynchronously.
func (s *APIKeyService) ValidateAndGetUserID(ctx context.Context, rawKey string) (string, error) {
	h := hash.SHA256(rawKey)
	k, err := s.keys.GetByHash(ctx, h)
	if err != nil {
		return "", domain.Unauthorized("invalid api key")
	}
	// Update last_used in background — don't fail the request
	go func() { _ = s.keys.UpdateLastUsed(context.Background(), k.ID) }()
	return k.UserID, nil
}

func (s *APIKeyService) ListByUser(ctx context.Context, userID string) ([]*domain.APIKey, error) {
	return s.keys.ListByUser(ctx, userID)
}

// Revoke archives an API key. Only the key's owner or an admin may revoke it.
func (s *APIKeyService) Revoke(ctx context.Context, userID, role, id string) error {
	key, err := s.keys.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key.UserID != userID && role != "admin" {
		return domain.Forbidden("cannot revoke another user's api key")
	}
	return s.keys.Archive(ctx, id)
}
