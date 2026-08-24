package centreon

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
)

func TestMediaService_List(t *testing.T) {
	mux, c := newTestMux(t)
	// The list representation uses name (not filename) and carries url; a swap of
	// the name/filename tags or a dropped url would fail here.
	mux.HandleFunc("GET /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 3, "name": "probe.png", "directory": "probe", "md5": "abc123", "url": "/centreon/img/media/probe/probe.png"},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Medias.List(t.Context())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	wantInt(t, "len(result)", len(resp.Result), 1)
	m := resp.Result[0]
	wantInt(t, "ID", m.ID, 3)
	wantStr(t, "Name", m.Name, "probe.png")
	wantStr(t, "Directory", m.Directory, "probe")
	wantStr(t, "MD5", m.MD5, "abc123")
	wantStr(t, "URL", m.URL, "/centreon/img/media/probe/probe.png")
}

func TestMediaService_Get(t *testing.T) {
	mux, c := newTestMux(t)
	// The detail representation uses filename (not name) and carries comment
	// instead of url; distinct from the list Media, so a shared struct would drop
	// one representation's fields.
	mux.HandleFunc("GET /centreon/api/latest/configuration/medias/3", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"id": 3, "comment": "an icon", "directory": "probe", "filename": "probe.png", "md5": "def456",
		})
	})

	got, err := c.Medias.Get(t.Context(), 3)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	wantInt(t, "ID", got.ID, 3)
	wantStr(t, "Comment", got.Comment, "an icon")
	wantStr(t, "Directory", got.Directory, "probe")
	wantStr(t, "Filename", got.Filename, "probe.png")
	wantStr(t, "MD5", got.MD5, "def456")
}

func TestMediaService_Create(t *testing.T) {
	mux, c := newTestMux(t)
	var (
		gotContentType string
		gotFilename    string
		gotDirectory   string
		gotData        []byte
	)
	mux.HandleFunc("POST /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		//nolint:gosec // G120: test server handler parsing a tiny, known multipart body under our control.
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm: %v", err)
		}
		// Read the directory from the already-parsed form rather than FormValue,
		// which would trigger another (unbounded) parse.
		if vs := r.MultipartForm.Value["directory"]; len(vs) > 0 {
			gotDirectory = vs[0]
		}
		f, hdr, err := r.FormFile("data")
		if err != nil {
			t.Errorf("FormFile(data): %v", err)
		} else {
			gotFilename = hdr.Filename
			gotData, _ = io.ReadAll(f)
			_ = f.Close()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{"id": 7, "filename": "server.png", "directory": "icons", "md5": "aabbcc"}},
			"errors": []any{},
		})
	})

	res, err := c.Medias.Create(t.Context(), &CreateMediaRequest{
		Filename:  "server.png",
		Directory: "icons",
		Data:      []byte("\x89PNG\r\n\x1a\nfake-image-bytes"),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// The request must be multipart, with the image as a file part named "data"
	// and directory as a form field. Dropping the Data field, or sending JSON
	// instead of multipart, fails these assertions.
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("Content-Type = %q, want multipart/form-data", gotContentType)
	}
	wantStr(t, "sent filename", gotFilename, "server.png")
	wantStr(t, "sent directory", gotDirectory, "icons")
	if string(gotData) != "\x89PNG\r\n\x1a\nfake-image-bytes" {
		t.Errorf("sent data = %q, want the raw image bytes", gotData)
	}
	// The created media is returned from the batch result[0].
	wantInt(t, "result.ID", res.ID, 7)
	wantStr(t, "result.Filename", res.Filename, "server.png")
	wantStr(t, "result.Directory", res.Directory, "icons")
	wantStr(t, "result.MD5", res.MD5, "aabbcc")
}

func TestMediaService_Create_EmptyResult(t *testing.T) {
	mux, c := newTestMux(t)
	// A 200 with neither a result nor an error must not be reported as success:
	// returning &resp.Result[0] would panic. The empty-result guard turns it into
	// an error.
	mux.HandleFunc("POST /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"result": []any{}, "errors": []any{}})
	})

	res, err := c.Medias.Create(t.Context(), &CreateMediaRequest{Filename: "x.png", Directory: "d", Data: []byte("x")})
	if err == nil {
		t.Fatal("expected an error on an empty result, got nil")
	}
	if res != nil {
		t.Errorf("result = %v, want nil on error", res)
	}
	if !strings.Contains(err.Error(), "empty result") {
		t.Errorf("error = %v, want it to mention an empty result", err)
	}
}

func TestMediaService_Get_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/medias/9", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"code": 500, "message": "boom"})
	})

	got, err := c.Medias.Get(t.Context(), 9)
	if err == nil {
		t.Fatal("expected an error on HTTP 500, got nil")
	}
	if got != nil {
		t.Errorf("result = %v, want nil on error", got)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

func TestMediaService_All(t *testing.T) {
	mux, c := newTestMux(t)
	// One full page (total equals the page length), so the iterator yields the
	// element and then stops. Pins the All wrapper on the default (non-integration)
	// test path.
	mux.HandleFunc("GET /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{"id": 1, "name": "a.png", "directory": "d", "md5": "m", "url": "/u"}},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	var ids []int
	for m, err := range c.Medias.All(t.Context()) {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		ids = append(ids, m.ID)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Errorf("All yielded ids %v, want [1]", ids)
	}
}

func TestMediaService_AllFolders(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("GET /centreon/api/latest/configuration/media/folders", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{"id": 2, "name": "probe", "alias": "probe", "comment": nil}},
			"meta":   map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	var names []string
	for f, err := range c.Medias.AllFolders(t.Context()) {
		if err != nil {
			t.Fatalf("AllFolders: %v", err)
		}
		names = append(names, f.Name)
	}
	if len(names) != 1 || names[0] != "probe" {
		t.Errorf("AllFolders yielded names %v, want [probe]", names)
	}
}

func TestMediaService_Create_ServerErrors(t *testing.T) {
	mux, c := newTestMux(t)
	// A 200 batch response can still carry per-file errors; Create must surface
	// them rather than returning a bogus success.
	mux.HandleFunc("POST /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []any{},
			"errors": []string{"directory is invalid"},
		})
	})

	res, err := c.Medias.Create(t.Context(), &CreateMediaRequest{Filename: "x.png", Directory: "bad", Data: []byte("x")})
	if err == nil {
		t.Fatal("expected an error when the batch reports errors, got nil")
	}
	if res != nil {
		t.Errorf("result = %v, want nil on error", res)
	}
	if !strings.Contains(err.Error(), "directory is invalid") {
		t.Errorf("error = %v, want it to include the server error detail", err)
	}
}

func TestMediaService_Create_Error(t *testing.T) {
	mux, c := newTestMux(t)
	mux.HandleFunc("POST /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "[data] The property data is required"})
	})

	res, err := c.Medias.Create(t.Context(), &CreateMediaRequest{Filename: "x.png", Directory: "d", Data: []byte("x")})
	if err == nil {
		t.Fatal("expected an error on HTTP 400, got nil")
	}
	if res != nil {
		t.Errorf("result = %v, want nil on error", res)
	}
	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Errorf("expected *APIError, got %T", err)
	}
}

// TestMediaService_Create_401Retry proves the multipart body (carried as a
// *rawBody) is replayed correctly when a 401 triggers re-authentication: the
// bytes are re-read from a fresh reader on the retry, so the second request
// still carries the full image.
func TestMediaService_Create_401Retry(t *testing.T) {
	mux, c := newTestMux(t)
	c.username = "admin"
	c.password = "secret"

	var calls atomic.Int32
	var replayedData []byte
	mux.HandleFunc("POST /centreon/api/latest/login", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, loginResponse{Security: loginSecurityResponse{Token: "fresh-token"}})
	})
	mux.HandleFunc("POST /centreon/api/latest/configuration/medias", func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		//nolint:gosec // G120: test server handler parsing a tiny, known multipart body under our control.
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("ParseMultipartForm on retry: %v", err)
		}
		if f, _, err := r.FormFile("data"); err == nil {
			replayedData, _ = io.ReadAll(f)
			_ = f.Close()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{{"id": 9, "filename": "r.png", "directory": "d", "md5": "z"}},
			"errors": []any{},
		})
	})

	res, err := c.Medias.Create(t.Context(), &CreateMediaRequest{Filename: "r.png", Directory: "d", Data: []byte("retry-image-bytes")})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("create called %d times, want 2 (401 then retry)", calls.Load())
	}
	if string(replayedData) != "retry-image-bytes" {
		t.Errorf("replayed data = %q, want the full image bytes re-sent on retry", replayedData)
	}
	wantInt(t, "result.ID", res.ID, 9)
}

func TestMediaService_Delete(t *testing.T) {
	mux, c := newTestMux(t)
	var called bool
	mux.HandleFunc("DELETE /centreon/api/latest/configuration/medias/5", func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.Medias.Delete(t.Context(), 5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !called {
		t.Error("expected DELETE /configuration/medias/5 to be called")
	}
}

func TestMediaService_ListFolders(t *testing.T) {
	mux, c := newTestMux(t)
	// The folders route is singular: /configuration/media/folders (not medias).
	// Comment is nullable, so it must decode to a nil *string.
	mux.HandleFunc("GET /centreon/api/latest/configuration/media/folders", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"result": []map[string]any{
				{"id": 2, "name": "probe", "alias": "probe", "comment": nil},
			},
			"meta": map[string]any{"page": 1, "limit": 10, "total": 1},
		})
	})

	resp, err := c.Medias.ListFolders(t.Context())
	if err != nil {
		t.Fatalf("ListFolders: %v", err)
	}
	wantInt(t, "len(result)", len(resp.Result), 1)
	f := resp.Result[0]
	wantInt(t, "ID", f.ID, 2)
	wantStr(t, "Name", f.Name, "probe")
	wantStr(t, "Alias", f.Alias, "probe")
	wantNilStrPtr(t, "Comment", f.Comment)
}

// TestMedia_WireTags pins the exact JSON keys of every media struct. The list
// Media and detail MediaDetail deliberately differ (name+url vs filename+
// comment); a copy-paste that shared the wrong tag set would pass a decode test
// but fail here.
func TestMedia_WireTags(t *testing.T) {
	comment := "c"
	cases := []struct {
		name string
		v    any
		want string
	}{
		{
			"Media",
			Media{ID: 1, Name: "n", Directory: "d", MD5: "m", URL: "u"},
			`{"id":1,"name":"n","directory":"d","md5":"m","url":"u"}`,
		},
		{
			"MediaDetail",
			MediaDetail{ID: 1, Comment: "c", Directory: "d", Filename: "f", MD5: "m"},
			`{"id":1,"comment":"c","directory":"d","filename":"f","md5":"m"}`,
		},
		{
			"MediaFolder",
			MediaFolder{ID: 1, Name: "n", Alias: "a", Comment: &comment},
			`{"id":1,"name":"n","alias":"a","comment":"c"}`,
		},
		{
			"MediaCreateResult",
			MediaCreateResult{ID: 1, Filename: "f", Directory: "d", MD5: "m"},
			`{"id":1,"filename":"f","directory":"d","md5":"m"}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.v)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("Marshal(%s) = %s, want %s", tc.name, got, tc.want)
			}
		})
	}
}
