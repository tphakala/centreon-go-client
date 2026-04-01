package centreon

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestQueryParams_Empty(t *testing.T) {
	opts := &ListOptions{}
	q := opts.queryParams()
	if len(q) != 0 {
		t.Errorf("expected empty query params, got %v", q)
	}
}

func TestQueryParams_Page(t *testing.T) {
	opts := &ListOptions{Page: 2, Limit: 50}
	q := opts.queryParams()
	if q.Get("page") != "2" {
		t.Errorf("page = %q, want %q", q.Get("page"), "2")
	}
	if q.Get("limit") != "50" {
		t.Errorf("limit = %q, want %q", q.Get("limit"), "50")
	}
}

func TestQueryParams_Search(t *testing.T) {
	opts := &ListOptions{Search: Eq("name", "host1")}
	q := opts.queryParams()
	got := q.Get("search")
	want := `{"name":{"$eq":"host1"}}`
	if got != want {
		t.Errorf("search = %q, want %q", got, want)
	}
}

func TestQueryParams_Sort(t *testing.T) {
	opts := &ListOptions{SortBy: map[string]string{"name": "asc"}}
	q := opts.queryParams()
	got := q.Get("sort_by")
	// Parse to avoid map ordering issues
	var m map[string]string
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("unmarshal sort_by: %v", err)
	}
	if m["name"] != "asc" {
		t.Errorf("sort_by[name] = %q, want %q", m["name"], "asc")
	}
}

func TestWithPage(t *testing.T) {
	opts := &ListOptions{}
	WithPage(3)(opts)
	if opts.Page != 3 {
		t.Errorf("Page = %d, want 3", opts.Page)
	}
}

func TestWithLimit(t *testing.T) {
	opts := &ListOptions{}
	WithLimit(100)(opts)
	if opts.Limit != 100 {
		t.Errorf("Limit = %d, want 100", opts.Limit)
	}
}

func TestWithSearch(t *testing.T) {
	f := Eq("name", "x")
	opts := &ListOptions{}
	WithSearch(f)(opts)
	if opts.Search == nil {
		t.Fatal("Search should not be nil")
	}
}

func TestWithSort(t *testing.T) {
	opts := &ListOptions{}
	WithSort(map[string]string{"id": "desc"})(opts)
	if opts.SortBy["id"] != "desc" {
		t.Errorf("SortBy[id] = %q, want %q", opts.SortBy["id"], "desc")
	}
}

func TestClientList_Integration(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/hosts", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page != "1" {
			t.Errorf("page = %q, want %q", page, "1")
		}
		writeJSON(w, 200, map[string]any{
			"result": []item{{ID: 1, Name: "host1"}},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	var resp ListResponse[item]
	err := c.list(t.Context(), "/hosts", []ListOption{WithPage(1), WithLimit(10)}, &resp)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(resp.Result) != 1 {
		t.Fatalf("len(Result) = %d, want 1", len(resp.Result))
	}
	if resp.Result[0].Name != "host1" {
		t.Errorf("Name = %q, want %q", resp.Result[0].Name, "host1")
	}
	if resp.Meta.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Meta.Total)
	}
}

func TestAll_TwoPages(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	mux, c := newTestMux(t)
	callCount := 0

	mux.HandleFunc("GET /centreon/api/latest/items", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			writeJSON(w, 200, map[string]any{
				"result": []item{{ID: 1}, {ID: 2}},
				"meta":   map[string]any{"page": 1, "limit": 2, "total": 3},
			})
		case "2":
			writeJSON(w, 200, map[string]any{
				"result": []item{{ID: 3}},
				"meta":   map[string]any{"page": 2, "limit": 2, "total": 3},
			})
		default:
			t.Errorf("unexpected page %q", page)
		}
	})

	listFn := func(ctx context.Context, opts ...ListOption) (*ListResponse[item], error) {
		var resp ListResponse[item]
		err := c.list(ctx, "/items", opts, &resp)
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	var collected []int
	for ptr, err := range all(t.Context(), listFn, []ListOption{WithLimit(2)}) {
		if err != nil {
			t.Fatalf("all: %v", err)
		}
		collected = append(collected, ptr.ID)
	}

	if len(collected) != 3 {
		t.Fatalf("collected %d items, want 3", len(collected))
	}
	for i, want := range []int{1, 2, 3} {
		if collected[i] != want {
			t.Errorf("collected[%d] = %d, want %d", i, collected[i], want)
		}
	}
	if callCount != 2 {
		t.Errorf("handler called %d times, want 2", callCount)
	}
}

func TestGetByID_Found(t *testing.T) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/items", func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		expected := `{"id":{"$eq":42}}`
		if search != expected {
			t.Errorf("search = %q, want %q", search, expected)
		}
		writeJSON(w, 200, map[string]any{
			"result": []item{{ID: 42, Name: "found"}},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	listFn := func(ctx context.Context, opts ...ListOption) (*ListResponse[item], error) {
		var resp ListResponse[item]
		err := c.list(ctx, "/items", opts, &resp)
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	got, err := getByID(t.Context(), listFn, "item", 42)
	if err != nil {
		t.Fatalf("getByID: %v", err)
	}
	if got.ID != 42 || got.Name != "found" {
		t.Errorf("got ID=%d Name=%q, want 42/found", got.ID, got.Name)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	mux, c := newTestMux(t)

	mux.HandleFunc("GET /centreon/api/latest/items", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{
			"result": []item{},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 0},
		})
	})

	listFn := func(ctx context.Context, opts ...ListOption) (*ListResponse[item], error) {
		var resp ListResponse[item]
		err := c.list(ctx, "/items", opts, &resp)
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	_, err := getByID(t.Context(), listFn, "item", 99)
	if err == nil {
		t.Fatal("expected error")
	}
	nfErr, ok := errors.AsType[*NotFoundError](err)
	if !ok {
		t.Fatalf("expected *NotFoundError, got %T: %v", err, err)
	}
	if nfErr.Resource != "item" || nfErr.ID != 99 {
		t.Errorf("got Resource=%q ID=%d, want item/99", nfErr.Resource, nfErr.ID)
	}
}

// TestAll_BreakEarly verifies that the iterator stops fetching when the
// consumer breaks out of the range loop.
func TestAll_BreakEarly(t *testing.T) {
	type item struct {
		ID int `json:"id"`
	}

	mux, c := newTestMux(t)
	callCount := 0

	mux.HandleFunc("GET /centreon/api/latest/items", func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		writeJSON(w, 200, map[string]any{
			"result": []item{{ID: callCount}},
			"meta":   map[string]any{"page": callCount, "limit": 1, "total": 100},
		})
	})

	listFn := func(ctx context.Context, opts ...ListOption) (*ListResponse[item], error) {
		var resp ListResponse[item]
		err := c.list(ctx, "/items", opts, &resp)
		if err != nil {
			return nil, err
		}
		return &resp, nil
	}

	count := 0
	for _, err := range all(t.Context(), listFn, []ListOption{WithLimit(1)}) {
		if err != nil {
			t.Fatalf("all: %v", err)
		}
		count++
		if count >= 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	// Should have stopped after 2 pages, not fetched all 100
	if callCount > 3 {
		t.Errorf("handler called %d times, expected at most 3", callCount)
	}
}
