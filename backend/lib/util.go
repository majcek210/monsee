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
func ParseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(s)
	return u, err
}

func Int32Ptr(i int) *int32 {
	if i == 0 {
		return nil
	}
	v := int32(i)
	return &v
}
