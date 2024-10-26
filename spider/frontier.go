package spider

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"sync"
)

type Frontier struct {
	db          *sql.DB
	initialized bool
	mutex       sync.Mutex
}

func (f *Frontier) Init() {
	dbName := "frontier.db"
	exists, err := f.fileExists(dbName)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Println("frontier - Creating SQLite database for frontier...")
		file, err := os.Create("frontier.db")
		if err != nil {
			log.Fatal(err)
		}
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}

	f.db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}

	f.initialized = true
	f.createTable()
}

func (f *Frontier) createTable() {
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	log.Println("frontier - Creating table frontier and indexes...")
	createDB := `CREATE TABLE IF NOT EXISTS frontier (
					url TEXT PRIMARY KEY,
					goroutine INT NOT NULL
				 );
				 CREATE INDEX idx_goroutines
				 ON frontier (goroutine);`
	statement, err := f.db.Prepare(createDB)
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func (f *Frontier) PopURL(routineNum int) string {
	// Query and return a URL from the frontier DB
	// true if frontier isn't empty, false otherwise
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	var url string
	query := fmt.Sprintf("SELECT url FROM frontier WHERE goroutine = %d LIMIT 1", routineNum)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	result := f.db.QueryRow(query)
	err := result.Scan(&url)
	if err != nil {
		return ""
	}

	query = fmt.Sprintf(`DELETE FROM frontier WHERE url = '%s';`, url)
	statement, err := f.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}

	return url
}

func (f *Frontier) CheckURLInFrontier(url string) bool {
	// This function should only be used for debugging purposes,
	// it's much faster to check if a page has been downloaded
	// by checking if the hash exists as JSON form in the
	// pages directory
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	var exists bool
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM frontier WHERE url = '%s');`, url)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	result := f.db.QueryRow(query)
	err := result.Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	return exists
}

func (f *Frontier) InsertPage(url string, routineNum int) {
	if !f.initialized {
		log.Fatal("Must initialize database connection before operating on it")
	}
	query := fmt.Sprintf(`INSERT OR IGNORE INTO frontier (url, goroutine) VALUES ('%s', %d);`, url, routineNum)
	f.mutex.Lock()
	defer f.mutex.Unlock()
	statement, err := f.db.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	// Change to non-fatal log to prevent crashing
	_, err = statement.Exec()
	if err != nil {
		log.Println(err)
	}
	return
}

func (f *Frontier) fileExists(path string) (bool, error) {
	// Check if file exists on disk
	// Taken from: https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}
