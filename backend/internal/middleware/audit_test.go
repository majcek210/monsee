package middleware

import (
	"reflect"
	"testing"
)

func TestAuditFieldNamesSortedKeysOnly(t *testing.T) {
	body := []byte(`{"name":"my webhook","url":"https://evil.example/secret","secret":"sssh","enabled":true}`)

	got := auditFieldNames(body)
	want := []string{"enabled", "name", "secret", "url"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("auditFieldNames(%s) = %v, want %v", body, got, want)
	}

	for _, field := range got {
		if field == "https://evil.example/secret" || field == "sssh" {
			t.Fatalf("auditFieldNames leaked a value: %v", got)
		}
	}
}

func TestAuditFieldNamesEmptyBody(t *testing.T) {
	if got := auditFieldNames(nil); got != nil {
		t.Fatalf("auditFieldNames(nil) = %v, want nil", got)
	}
	if got := auditFieldNames([]byte{}); got != nil {
		t.Fatalf("auditFieldNames(empty) = %v, want nil", got)
	}
}

func TestAuditFieldNamesInvalidJSON(t *testing.T) {
	if got := auditFieldNames([]byte("not json")); got != nil {
		t.Fatalf("auditFieldNames(invalid) = %v, want nil", got)
	}
}

func TestAuditFieldNamesNonObjectJSON(t *testing.T) {
	if got := auditFieldNames([]byte(`[1,2,3]`)); got != nil {
		t.Fatalf("auditFieldNames(array) = %v, want nil", got)
	}
}
