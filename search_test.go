package centreon

import (
	"encoding/json"
	"testing"
)

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(data)
}

func TestEq(t *testing.T) {
	got := mustJSON(t, Eq("name", "host1").Build())
	want := `{"name":{"$eq":"host1"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNeq(t *testing.T) {
	got := mustJSON(t, Neq("name", "host1").Build())
	want := `{"name":{"$neq":"host1"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLt(t *testing.T) {
	got := mustJSON(t, Lt("id", 10).Build())
	want := `{"id":{"$lt":10}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLe(t *testing.T) {
	got := mustJSON(t, Le("id", 10).Build())
	want := `{"id":{"$le":10}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestGt(t *testing.T) {
	got := mustJSON(t, Gt("id", 10).Build())
	want := `{"id":{"$gt":10}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestGe(t *testing.T) {
	got := mustJSON(t, Ge("id", 10).Build())
	want := `{"id":{"$ge":10}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestLk(t *testing.T) {
	got := mustJSON(t, Lk("name", "host%").Build())
	want := `{"name":{"$lk":"host%"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNk(t *testing.T) {
	got := mustJSON(t, Nk("name", "host%").Build())
	want := `{"name":{"$nk":"host%"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRg(t *testing.T) {
	got := mustJSON(t, Rg("name", "^host").Build())
	want := `{"name":{"$rg":"^host"}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestIn(t *testing.T) {
	got := mustJSON(t, In("id", 1, 2, 3).Build())
	want := `{"id":{"$in":[1,2,3]}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNi(t *testing.T) {
	got := mustJSON(t, Ni("id", 1, 2).Build())
	want := `{"id":{"$ni":[1,2]}}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestAnd(t *testing.T) {
	f := And(Eq("name", "host1"), Gt("id", 5))
	got := mustJSON(t, f.Build())
	want := `{"$and":[{"name":{"$eq":"host1"}},{"id":{"$gt":5}}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestOr(t *testing.T) {
	f := Or(Eq("name", "a"), Eq("name", "b"))
	got := mustJSON(t, f.Build())
	want := `{"$or":[{"name":{"$eq":"a"}},{"name":{"$eq":"b"}}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestNested_AndOr(t *testing.T) {
	f := And(
		Eq("status", "up"),
		Or(
			Lk("name", "web%"),
			Lk("name", "db%"),
		),
	)
	got := mustJSON(t, f.Build())
	want := `{"$and":[{"status":{"$eq":"up"}},{"$or":[{"name":{"$lk":"web%"}},{"name":{"$lk":"db%"}}]}]}`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
