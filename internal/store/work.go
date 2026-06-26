package store

import (
	"database/sql"
	"regexp"
	"strings"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

type WorkImage struct {
	ID        int
	WorkID    int
	URL       string
	Alt       string
	SortOrder int
}

type Work struct {
	ID          int
	Slug        string
	SortOrder   int
	Title       string
	WorkType    string
	Client      string
	Year        int
	Tools       string
	Description string
	Website     string
	CoverURL    string
	Link        string
	Images      []WorkImage
}

type WorkStore struct {
	db *sql.DB
}

func NewWorkStore(db *sql.DB) *WorkStore {
	return &WorkStore{db: db}
}

const cols = `id, slug, sort_order, title, type, client, year, tools, description, website, link, cover_url`

func scanWork(s interface{ Scan(...any) error }, w *Work) error {
	return s.Scan(&w.ID, &w.Slug, &w.SortOrder, &w.Title, &w.WorkType, &w.Client, &w.Year, &w.Tools, &w.Description, &w.Website, &w.Link, &w.CoverURL)
}

var sortCols = map[string]string{
	"title-asc":   "title ASC",
	"title-desc":  "title DESC",
	"type-asc":    "type ASC",
	"type-desc":   "type DESC",
	"client-asc":  "client ASC",
	"client-desc": "client DESC",
	"year-asc":    "year ASC",
	"year-desc":   "year DESC",
	"tools-asc":   "tools ASC",
	"tools-desc":  "tools DESC",
}

func (s *WorkStore) List(page, limit int) ([]Work, int, error) {
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM work").Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := s.db.Query("SELECT "+cols+" FROM work ORDER BY sort_order ASC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var works []Work
	for rows.Next() {
		var w Work
		if err := scanWork(rows, &w); err != nil {
			return nil, 0, err
		}
		works = append(works, w)
	}
	return works, total, rows.Err()
}

func (s *WorkStore) loadImages(workID int) ([]WorkImage, error) {
	rows, err := s.db.Query("SELECT id, work_id, url, alt, sort_order FROM work_images WHERE work_id = ? ORDER BY sort_order", workID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var imgs []WorkImage
	for rows.Next() {
		var img WorkImage
		if err := rows.Scan(&img.ID, &img.WorkID, &img.URL, &img.Alt, &img.SortOrder); err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
	}
	return imgs, rows.Err()
}

func (s *WorkStore) GetBySlug(slug string) (*Work, error) {
	var w Work
	err := scanWork(s.db.QueryRow("SELECT "+cols+" FROM work WHERE slug = ?", slug), &w)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.Images, err = s.loadImages(w.ID)
	return &w, err
}

func (s *WorkStore) GetByID(id int) (*Work, error) {
	var w Work
	err := scanWork(s.db.QueryRow("SELECT "+cols+" FROM work WHERE id = ?", id), &w)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	w.Images, err = s.loadImages(w.ID)
	return &w, err
}

// WorkQuery describes a filtered, paginated work listing. Zero-valued fields
// are unset; Tool matches a single tool within a work's comma-separated list.
type WorkQuery struct {
	Type   string
	Client string
	Year   string
	Tool   string
	Search string
	Sort   string
	Page   int
	Limit  int
}

func (s *WorkStore) Filter(q WorkQuery) ([]Work, int, error) {
	where := []string{}
	args := []any{}

	if q.Type != "" {
		where = append(where, "type = ?")
		args = append(args, q.Type)
	}
	if q.Client != "" {
		where = append(where, "client = ?")
		args = append(args, q.Client)
	}
	if q.Year != "" {
		where = append(where, "CAST(year AS TEXT) = ?")
		args = append(args, q.Year)
	}
	if q.Tool != "" {
		where = append(where, "(',' || tools || ',') LIKE ?")
		args = append(args, "%,"+q.Tool+",%")
	}
	if q.Search != "" {
		where = append(where, "(title LIKE ? OR client LIKE ? OR description LIKE ?)")
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
	}

	base := "FROM work"
	if len(where) > 0 {
		base += " WHERE " + strings.Join(where, " AND ")
	}

	orderBy := " ORDER BY sort_order ASC"
	if cl, ok := sortCols[q.Sort]; ok {
		orderBy = " ORDER BY " + cl
	}

	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (q.Page - 1) * q.Limit
	rows, err := s.db.Query("SELECT "+cols+" "+base+orderBy+" LIMIT ? OFFSET ?", append(args, q.Limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var works []Work
	for rows.Next() {
		var w Work
		if err := scanWork(rows, &w); err != nil {
			return nil, 0, err
		}
		works = append(works, w)
	}
	return works, total, rows.Err()
}

func (s *WorkStore) UniqueTypes() ([]string, error) {
	return uniqueCol(s.db, "type")
}

func (s *WorkStore) UniqueClients() ([]string, error) {
	return uniqueCol(s.db, "client")
}

func (s *WorkStore) UniqueYears() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT CAST(year AS TEXT) FROM work ORDER BY year DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *WorkStore) UniqueTools() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT tools FROM work ORDER BY tools")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := map[string]bool{}
	var out []string
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" && !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out, rows.Err()
}

func uniqueCol(db *sql.DB, col string) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT " + col + " FROM work ORDER BY " + col)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *WorkStore) Create(w Work) (int, error) {
	w.Slug = slugify(w.Title)
	res, err := s.db.Exec(
		"INSERT INTO work (slug, sort_order, title, type, client, year, tools, description, website, link, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		w.Slug, w.SortOrder, w.Title, w.WorkType, w.Client, w.Year, w.Tools, w.Description, w.Website, w.Link, w.CoverURL,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s *WorkStore) Update(w Work) error {
	_, err := s.db.Exec(
		"UPDATE work SET sort_order = ?, title = ?, type = ?, client = ?, year = ?, tools = ?, description = ?, website = ?, link = ?, cover_url = ? WHERE id = ?",
		w.SortOrder, w.Title, w.WorkType, w.Client, w.Year, w.Tools, w.Description, w.Website, w.Link, w.CoverURL, w.ID,
	)
	return err
}

func (s *WorkStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM work WHERE id = ?", id)
	return err
}
