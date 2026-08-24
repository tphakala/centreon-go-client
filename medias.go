package centreon

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"mime/multipart"
)

// Media is a media-library image as returned by the list endpoint
// (GET /configuration/medias). Note the list representation uses name and
// carries a url (the static path the web server serves the image from), whereas
// the per-id detail (MediaDetail) uses filename and carries a comment instead;
// the two representations genuinely differ, so they are separate types.
type Media struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Directory string `json:"directory"`
	MD5       string `json:"md5"`
	URL       string `json:"url"`
}

// MediaDetail is a media-library image as returned by the per-id detail endpoint
// (GET /configuration/medias/{id}). Unlike the list Media it uses filename (not
// name), includes comment, and omits url.
type MediaDetail struct {
	ID        int    `json:"id"`
	Comment   string `json:"comment"`
	Directory string `json:"directory"`
	Filename  string `json:"filename"`
	MD5       string `json:"md5"`
}

// MediaFolder is a media-library folder as returned by
// GET /configuration/media/folders (note the singular "media" in the path).
// Folders are created implicitly when a media is uploaded into a new directory;
// there is no folder create or delete route.
type MediaFolder struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Alias   string  `json:"alias"`
	Comment *string `json:"comment"`
}

// CreateMediaRequest is the input for uploading one media image. The create
// endpoint is multipart/form-data, not JSON: Data is sent as a file part named
// "data" with the given Filename, and Directory is sent as a form field.
// Directory is required (an empty directory is rejected with HTTP 400). The
// directory is created if it does not already exist.
type CreateMediaRequest struct {
	// Filename is the uploaded file's name (for example "server.png"); it becomes
	// the media's filename/name.
	Filename string
	// Directory is the target folder name. Required; must be non-empty.
	Directory string
	// Data is the raw image bytes (not base64-encoded; the client sends them as a
	// multipart file part).
	Data []byte
}

// MediaCreateResult is one created media as reported by the create response
// (POST /configuration/medias returns {result: [...], errors: [...]}).
type MediaCreateResult struct {
	ID        int    `json:"id"`
	Filename  string `json:"filename"`
	Directory string `json:"directory"`
	MD5       string `json:"md5"`
}

// createMediaResponse is the batch envelope returned by the create endpoint.
type createMediaResponse struct {
	Result []MediaCreateResult `json:"result"`
	Errors []string            `json:"errors"`
}

// MediaService provides access to the media (image) library under
// /configuration/medias. Uploading a media is the supported way to obtain an
// image an icon or severity can reference. Content retrieval is not exposed: the
// GET /configuration/medias/{id}/content route returns HTTP 405 on Centreon Web
// 25.10.16; the list Media.URL is the static path a client fetches the raw image
// from instead.
type MediaService struct {
	client *Client
}

// List returns a paginated list of medias.
func (s *MediaService) List(ctx context.Context, opts ...ListOption) (*ListResponse[Media], error) {
	var resp ListResponse[Media]
	err := s.client.list(ctx, "/configuration/medias", opts, &resp)
	return &resp, err
}

// All returns an iterator over all medias.
func (s *MediaService) All(ctx context.Context, opts ...ListOption) iter.Seq2[*Media, error] {
	return all(ctx, s.List, opts)
}

// Get returns the media with the given ID.
func (s *MediaService) Get(ctx context.Context, id int) (*MediaDetail, error) {
	var result MediaDetail
	if err := s.client.get(ctx, fmt.Sprintf("/configuration/medias/%d", id), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Create uploads one media image via multipart/form-data and returns the created
// media (id, filename, directory, md5). Directory is required. If the server
// reports a per-file error, Create returns it as an error.
func (s *MediaService) Create(ctx context.Context, req *CreateMediaRequest) (*MediaCreateResult, error) {
	body, err := buildMediaMultipart(req)
	if err != nil {
		return nil, err
	}
	var resp createMediaResponse
	if err := s.client.post(ctx, "/configuration/medias", body, &resp); err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("centreon: create media: %v", resp.Errors)
	}
	if len(resp.Result) == 0 {
		return nil, fmt.Errorf("centreon: create media: empty result")
	}
	return &resp.Result[0], nil
}

// Delete removes the media with the given ID.
func (s *MediaService) Delete(ctx context.Context, id int) error {
	return s.client.delete(ctx, fmt.Sprintf("/configuration/medias/%d", id))
}

// ListFolders returns a paginated list of media folders
// (GET /configuration/media/folders).
func (s *MediaService) ListFolders(ctx context.Context, opts ...ListOption) (*ListResponse[MediaFolder], error) {
	var resp ListResponse[MediaFolder]
	err := s.client.list(ctx, "/configuration/media/folders", opts, &resp)
	return &resp, err
}

// AllFolders returns an iterator over all media folders.
func (s *MediaService) AllFolders(ctx context.Context, opts ...ListOption) iter.Seq2[*MediaFolder, error] {
	return all(ctx, s.ListFolders, opts)
}

// buildMediaMultipart encodes a CreateMediaRequest into a multipart/form-data
// body: a "data" file part carrying the image bytes plus a "directory" form
// field. The returned *rawBody replays losslessly on a 401 retry (see rawBody).
func buildMediaMultipart(req *CreateMediaRequest) (*rawBody, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("data", req.Filename)
	if err != nil {
		return nil, fmt.Errorf("centreon: build media upload: %w", err)
	}
	if _, err := part.Write(req.Data); err != nil {
		return nil, fmt.Errorf("centreon: build media upload: %w", err)
	}
	if err := w.WriteField("directory", req.Directory); err != nil {
		return nil, fmt.Errorf("centreon: build media upload: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("centreon: build media upload: %w", err)
	}
	return &rawBody{contentType: w.FormDataContentType(), data: buf.Bytes()}, nil
}
