package history_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zaidejjo/ax/internal/history"
)

func TestOpen_CreatesDatabase(t *testing.T) {
	path := tempDB(t)
	s, err := history.Open(path)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	defer s.Close()
	defer os.Remove(path)
}

func TestInsert_And_List(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	entry := history.Entry{
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Accept": "application/json",
		},
		Body:      "",
		Status:    200,
		BodySize:  42,
		Duration:  150 * time.Millisecond,
		CreatedAt: time.Now(),
	}

	id, err := s.Insert(entry)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	entries, err := s.List(10, 0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got := entries[0]
	if got.Method != "GET" {
		t.Errorf("expected GET, got %s", got.Method)
	}
	if got.URL != "https://api.example.com/users" {
		t.Errorf("unexpected URL: %s", got.URL)
	}
	if got.Status != 200 {
		t.Errorf("expected 200, got %d", got.Status)
	}
	if got.BodySize != 42 {
		t.Errorf("expected 42, got %d", got.BodySize)
	}
	if got.Headers["Accept"] != "application/json" {
		t.Errorf("missing accept header")
	}
}

func TestList_OrderedByRecent(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	// Insert two entries with different timestamps.
	t1 := time.Now().Add(-1 * time.Hour)
	t2 := time.Now()

	id1, _ := s.Insert(history.Entry{
		Method:    "GET",
		URL:       "/first",
		Status:    200,
		CreatedAt: t1,
	})
	id2, _ := s.Insert(history.Entry{
		Method:    "POST",
		URL:       "/second",
		Status:    201,
		CreatedAt: t2,
	})

	entries, _ := s.List(10, 0)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Most recent first.
	if entries[0].ID != id2 {
		t.Errorf("expected most recent first, got ID %d", entries[0].ID)
	}
	if entries[1].ID != id1 {
		t.Errorf("expected oldest second, got ID %d", entries[1].ID)
	}
}

func TestGet_Existing(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	id, _ := s.Insert(history.Entry{
		Method: "DELETE",
		URL:    "https://api.example.com/items/1",
		Status: 204,
	})

	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Method != "DELETE" {
		t.Errorf("expected DELETE, got %s", got.Method)
	}
}

func TestGet_NonExistent(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	_, err := s.Get(99999)
	if err == nil {
		t.Fatal("expected error for non-existent entry")
	}
}

func TestDelete(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	id, _ := s.Insert(history.Entry{
		Method: "GET",
		URL:    "https://api.example.com/delete",
		Status: 200,
	})

	if err := s.Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Verify it's gone.
	entries, _ := s.List(10, 0)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestDelete_NonExistent(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	err := s.Delete(99999)
	if err == nil {
		t.Fatal("expected error for deleting non-existent entry")
	}
}

func TestList_Limit(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	for i := 0; i < 5; i++ {
		s.Insert(history.Entry{
			Method: "GET",
			URL:    "/item",
			Status: 200,
		})
	}

	entries, err := s.List(3, 0)
	if err != nil {
		t.Fatalf("list limited: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries with limit 3, got %d", len(entries))
	}
}

func TestList_Offset(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	// Insert a few, then list with offset to paginate.
	for i := 0; i < 5; i++ {
		s.Insert(history.Entry{
			Method: "GET",
			URL:    "/item",
			Status: 200,
		})
	}

	// Skip the first 3, get the last 2.
	entries, err := s.List(10, 3)
	if err != nil {
		t.Fatalf("list with offset: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries with offset 3, got %d", len(entries))
	}
}

func TestEmptyStore_ReturnsEmptyList(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	entries, err := s.List(10, 0)
	if err != nil {
		t.Fatalf("list from empty store: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d", len(entries))
	}
}

func TestEntry_Duration(t *testing.T) {
	s := openTempDB(t)
	defer s.Close()

	id, _ := s.Insert(history.Entry{
		Method:   "GET",
		URL:      "/slow",
		Status:   200,
		Duration: 2 * time.Second,
	})

	got, _ := s.Get(id)
	if got.Duration != 2*time.Second {
		t.Errorf("expected 2s, got %v", got.Duration)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func tempDB(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.db")
}

func openTempDB(t *testing.T) *history.Store {
	t.Helper()
	s, err := history.Open(tempDB(t))
	if err != nil {
		t.Fatalf("open temp db: %v", err)
	}
	return s
}
