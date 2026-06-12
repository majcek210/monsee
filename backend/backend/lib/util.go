package lib

import "github.com/jackc/pgx/v5/pgtype"

func PgtypeText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
