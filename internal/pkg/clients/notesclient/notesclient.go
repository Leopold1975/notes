package notesclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"notes/internal/pkg/config"
	"notes/internal/pkg/models"
)

var ErrNotFound = errors.New("not found")

var notesPath = "notes"

type NotesClient struct {
	client *http.Client
	cfg    config.Server
}

func New(cfg config.Server) NotesClient {
	return NotesClient{
		client: http.DefaultClient,
		cfg:    cfg,
	}
}

func (n *NotesClient) GetNoteRequest(id uint64) (models.Note, error) {
	note := models.Note{}

	s := strconv.Itoa(int(id))
	u := url.URL{
		Scheme: "http",
		Host:   n.cfg.Host + n.cfg.Port,
		Path:   path.Join(notesPath, s),
	}

	b, err := n.DoRequest(http.MethodGet, u.String())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return note, ErrNotFound
		}
		return note, fmt.Errorf("cannot do request to server error: %w", err)
	}

	err = json.Unmarshal(b, &note)
	if err != nil {
		return note, fmt.Errorf("cannot unmarshal server's response error: %w", err)
	}

	return note, err
}

func (n *NotesClient) GetNotesRequest() ([]models.Note, error) {
	u := url.URL{
		Scheme: "http",
		Host:   n.cfg.Host + n.cfg.Port,
		Path:   notesPath,
	}

	b, err := n.DoRequest(http.MethodGet, u.String())
	if err != nil {
		return nil, err
	}

	notes := make([]models.Note, 0, 4)
	if err := json.Unmarshal(b, &notes); err != nil {
		return nil, fmt.Errorf("cannot unmarshal server's response error: %w", err)
	}
	return notes, nil
}

func (n *NotesClient) DeleteNoteRequest(id uint64) error {
	q := url.Values{}
	q.Add("id", strconv.Itoa(int(id)))
	u := url.URL{
		Scheme:   "http",
		Host:     n.cfg.Host + n.cfg.Port,
		Path:     notesPath,
		RawQuery: q.Encode(),
	}

	fmt.Println(u.String())
	_, err := n.DoRequest(http.MethodDelete, u.String())
	if err != nil {
		return err
	}

	return nil
}

func (n *NotesClient) UpdateNoteRequest(_ models.Note) error {
	panic("not implemented")
	// return nil
}

func (n *NotesClient) CreateNoteRequest(_ models.Note) error {
	panic("not implemented")
	// return nil
}

func (n *NotesClient) DoRequest(method string, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return b, err
}
