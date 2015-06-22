package settings

var (
	Port string

	DB struct {
		Engine string
		Source string
	}
)

func init() {
	DB.Engine = "sqlite3"
	DB.Source = "/tmp/db.sqlite"
}
