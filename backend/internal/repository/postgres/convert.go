package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/majcek210/monsee/lib"
)

func uuidStr(u pgtype.UUID) string {
	return uuid.UUID(u.Bytes).String()
}

func uuidStrPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func parseUUID(s string) (pgtype.UUID, error) {
	return lib.ParseUUID(s)
}

func mustParseUUID(s string) pgtype.UUID {
	u, err := lib.ParseUUID(s)
	if err != nil {
		panic("invalid UUID: " + s)
	}
	return u
}

func parseUUIDPtr(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{}
	}
	u, err := lib.ParseUUID(*s)
	if err != nil {
		return pgtype.UUID{}
	}
	return u
}

func tsToTime(ts pgtype.Timestamptz) time.Time {
	return ts.Time
}

func tsToTimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func timeToPgtz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
