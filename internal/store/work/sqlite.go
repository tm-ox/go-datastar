package work

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

type SQLiteWorkStore struct {
	db *sql.DB
}

func NewSQLiteWorkStore(db *sql.DB) *SQLiteWorkStore {
	return &SQLiteWorkStore{db: db}
}

const cols = `id, slug, title, type, client, year, tools, description, website, link, cover_url`

func scanWork(s interface{ Scan(...any) error }, w *Work) error {
	return s.Scan(&w.ID, &w.Slug, &w.Title, &w.WorkType, &w.Client, &w.Year, &w.Tools, &w.Description, &w.Website, &w.Link, &w.CoverURL)
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

func (s *SQLiteWorkStore) List(page, limit int) ([]Work, int, error) {
	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM work").Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := s.db.Query("SELECT "+cols+" FROM work ORDER BY year DESC LIMIT ? OFFSET ?", limit, offset)
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

func (s *SQLiteWorkStore) loadImages(workID int) ([]WorkImage, error) {
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

func (s *SQLiteWorkStore) GetBySlug(slug string) (*Work, error) {
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

func (s *SQLiteWorkStore) GetByID(id int) (*Work, error) {
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

func (s *SQLiteWorkStore) Filter(workType, client, year, tools, search, sort string, page, limit int) ([]Work, int, error) {
	where := []string{}
	args := []any{}

	if workType != "" {
		where = append(where, "type = ?")
		args = append(args, workType)
	}
	if client != "" {
		where = append(where, "client = ?")
		args = append(args, client)
	}
	if year != "" {
		where = append(where, "CAST(year AS TEXT) = ?")
		args = append(args, year)
	}
	if tools != "" {
		where = append(where, "(',' || tools || ',') LIKE ?")
		args = append(args, "%,"+tools+",%")
	}
	if search != "" {
		where = append(where, "(title LIKE ? OR client LIKE ? OR description LIKE ?)")
		args = append(args, "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	base := "FROM work"
	if len(where) > 0 {
		base += " WHERE " + strings.Join(where, " AND ")
	}

	orderBy := " ORDER BY year DESC"
	if cl, ok := sortCols[sort]; ok {
		orderBy = " ORDER BY " + cl
	}

	var total int
	if err := s.db.QueryRow("SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	rows, err := s.db.Query("SELECT "+cols+" "+base+orderBy+" LIMIT ? OFFSET ?", append(args, limit, offset)...)
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

func (s *SQLiteWorkStore) UniqueTypes() ([]string, error) {
	return uniqueCol(s.db, "type")
}

func (s *SQLiteWorkStore) UniqueClients() ([]string, error) {
	return uniqueCol(s.db, "client")
}

func (s *SQLiteWorkStore) UniqueYears() ([]string, error) {
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

func (s *SQLiteWorkStore) UniqueTools() ([]string, error) {
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

func (s *SQLiteWorkStore) Create(w Work) (int, error) {
	w.Slug = slugify(w.Title)
	res, err := s.db.Exec(
		"INSERT INTO work (slug, title, type, client, year, tools, description, website, link, cover_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		w.Slug, w.Title, w.WorkType, w.Client, w.Year, w.Tools, w.Description, w.Website, w.Link, w.CoverURL,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return int(id), err
}

func (s *SQLiteWorkStore) Update(w Work) error {
	_, err := s.db.Exec(
		"UPDATE work SET title = ?, type = ?, client = ?, year = ?, tools = ?, description = ?, website = ?, link = ?, cover_url = ? WHERE id = ?",
		w.Title, w.WorkType, w.Client, w.Year, w.Tools, w.Description, w.Website, w.Link, w.CoverURL, w.ID,
	)
	return err
}

func (s *SQLiteWorkStore) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM work WHERE id = ?", id)
	return err
}
