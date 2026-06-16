package work

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

type WorkStore interface {
	List(page, limit int) ([]Work, int, error)
	GetBySlug(slug string) (*Work, error)
	GetByID(id int) (*Work, error)
	Filter(workType, client, year, tools, search, sort string, page, limit int) ([]Work, int, error)
	UniqueTypes() ([]string, error)
	UniqueClients() ([]string, error)
	UniqueYears() ([]string, error)
	UniqueTools() ([]string, error)
	Create(w Work) (int, error)
	Update(w Work) error
	Delete(id int) error
}
