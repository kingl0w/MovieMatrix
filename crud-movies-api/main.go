package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

const defaultCoverURL = "https://www.reelviews.net/resources/img/default_poster.jpg"

type Movie struct {
	ID string `json:"id"`
	MID string `json:"mid"`
	Title string `json:"title"`
	Director Director `json:"director"`
    Cover string `json:"cover"`
    Categories []Category `json:"categories"`
}

type Director struct {
	ID int `json:"id"`
	Firstname string `json:"firstname"`
	Lastname string `json:"lastname"`
}

type Category struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

var db *sql.DB

func init() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func generateMID(movie Movie) string {
    info := movie.Title + movie.Director.Firstname + movie.Director.Lastname

    for _, category := range movie.Categories {
        info += category.Name
    }

    hash := sha256.Sum256([]byte(info))

    mid := hex.EncodeToString(hash[:10])
    
    return mid[:4] + "-" + mid[4:8] + "-" + mid[8:10]
}

func validateMovie(movie Movie) error {
    if movie.Title == "" {
        return fmt.Errorf("title is required")
    }
    if movie.Director.Firstname == "" || movie.Director.Lastname == "" {
        return fmt.Errorf("director's first name and last name are required")
    }
    if movie.Cover == "" {
        movie.Cover = defaultCoverURL
    }
    return nil
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
    err := db.Ping()
    if err != nil {
        http.Error(w, "Database connection failed", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}
func initDB() {
	var err error
	
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_PORT"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"))

	log.Println("Connecting to database...")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	} 

	log.Println("Pinging database...")
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Creating tables if not exist...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS directors (
			id SERIAL PRIMARY KEY,
			firstname TEXT,
			lastname TEXT
			)`)
		if err != nil {
			log.Fatal(err)
		}

		_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS movies (
            id SERIAL PRIMARY KEY,
            mid TEXT,
            title TEXT,
            director_id INTEGER REFERENCES directors(id),
            cover TEXT
            )`)
		if err != nil {
			log.Fatal(err)
		}

        _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS categories (
            id SERIAL PRIMARY KEY,
            name TEXT UNIQUE
            )`)
        if err != nil {
            log.Fatal(err)
        }

        _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS movie_categories (
            movie_id INTEGER REFERENCES movies(id),
            category_id INTEGER REFERENCES categories(id),
            PRIMARY KEY (movie_id, category_id)
            )`)
        if err != nil {
            log.Fatal(err)
        }
	    log.Println("Database initialization complete")
}

func getMovies(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    if id, ok := params["id"]; ok {
        //Get single movie
        var movie Movie
        err := db.QueryRow(`
            SELECT m.id, m.mid, m.title, m.cover, d.id, d.firstname, d.lastname 
            FROM movies m 
            JOIN directors d ON m.director_id = d.id 
            WHERE m.id = $1`, id).Scan(
            &movie.ID, &movie.MID, &movie.Title, &movie.Cover, &movie.Director.ID, &movie.Director.Firstname, &movie.Director.Lastname)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "Movie not found", http.StatusNotFound)
            } else {
                http.Error(w, err.Error(), http.StatusInternalServerError)
            }
            return
        }

        rows, err := db.Query(`
            SELECT c.id, c.name 
            FROM categories c 
            JOIN movie_categories mc ON c.id = mc.category_id 
            WHERE mc.movie_id = $1`, id)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()
        
        for rows.Next() {
            var category Category
            if err := rows.Scan(&category.ID, &category.Name); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            movie.Categories = append(movie.Categories, category)
        }

        json.NewEncoder(w).Encode(movie)
    } else {
        //Get all movies
        rows, err := db.Query(`
        SELECT m.id, m.mid, m.title, m.cover, d.id, d.firstname, d.lastname
        FROM movies m
        JOIN directors d ON m.director_id = d.id
        Order BY m.title ASC`)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var movies []Movie
    for rows.Next() {
        var m Movie
        err := rows.Scan(&m.ID, &m.MID, &m.Title, &m.Cover, &m.Director.ID, &m.Director.Firstname, &m.Director.Lastname)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Fetch categories for each movie
        catRows, err := db.Query(`
            SELECT c.id, c.name 
            FROM categories c 
            JOIN movie_categories mc ON c.id = mc.category_id 
            WHERE mc.movie_id = $1`, m.ID)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer catRows.Close()
        
        for catRows.Next() {
            var category Category
            if err := catRows.Scan(&category.ID, &category.Name); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            m.Categories = append(m.Categories, category)
        }
        
        movies = append(movies, m)
    }
    json.NewEncoder(w).Encode(movies)
    }

}

func searchMovies(w http.ResponseWriter, r *http.Request) {
    log.Println("searchMovies function called")

    if mux.Vars(r)["id"] != "" {
        log.Println("Invalid search request: id parameter detected")
        http.Error(w, "Invalid search request", http.StatusBadRequest)
        return
    }

    query := r.URL.Query().Get("q")
    if query == "" {
        log.Println("Empty search query")
        http.Error(w, "Search query is required", http.StatusBadRequest)
        return
    }

    log.Printf("Searching for movies with query: %s", query)
    likeQuery := "%" + query + "%"

    if err := db.Ping(); err != nil {
        log.Printf("Database connection error: %v", err)
        http.Error(w, "Database connection error", http.StatusInternalServerError)
        return
    }
    log.Println("Database connection successful")

    rows, err := db.Query(`
        SELECT DISTINCT m.id, m.mid, m.title, m.cover, d.id, d.firstname, d.lastname
        FROM movies m
        JOIN directors d ON m.director_id = d.id
        LEFT JOIN movie_categories mc ON m.id = mc.movie_id
        LEFT JOIN categories c ON mc.category_id = c.id
        WHERE m.title ILIKE $1
        OR d.firstname ILIKE $1
        OR d.lastname ILIKE $1
        OR c.name ILIKE $1
    `, likeQuery)

    if err != nil {
        log.Printf("Error executing search query: %v", err)
        http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    log.Println("Query executed successfully")

    var movies []Movie
    for rows.Next() {
        var m Movie
        err := rows.Scan(&m.ID, &m.MID, &m.Title, &m.Cover, &m.Director.ID, &m.Director.Firstname, &m.Director.Lastname)
        if err != nil {
            log.Printf("Error scanning row: %v", err)
            http.Error(w, fmt.Sprintf("Error scanning row: %v", err), http.StatusInternalServerError)
            return
        }
        log.Printf("Scanned movie: %+v", m)
        movies = append(movies, m)
    }

    if err = rows.Err(); err != nil {
        log.Printf("Error after scanning rows: %v", err)
        http.Error(w, fmt.Sprintf("Error after scanning rows: %v", err), http.StatusInternalServerError)
        return
    }

    log.Printf("Found %d movies matching query: %s", len(movies), query)

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(movies); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Error encoding response", http.StatusInternalServerError)
        return
    }

    log.Println("Response sent successfully")
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
    log.Println("updateMovie function called")
    params := mux.Vars(r)
    id := params["id"]
    log.Printf("Updating movie with ID: %s", id)

    var movie Movie
    err := json.NewDecoder(r.Body).Decode(&movie)
    if err != nil {
        log.Printf("Error decoding request body: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    log.Printf("Received movie data: %+v", movie)

    if err := validateMovie(movie); err != nil {
        log.Printf("Movie validation failed: %v", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    tx, err := db.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update movie
    _, err = tx.Exec("UPDATE movies SET mid = $1, title = $2, cover = $3 WHERE id = $4",
        movie.MID, movie.Title, movie.Cover, id)
    if err != nil {
        tx.Rollback()
        log.Printf("Error updating movie: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update director
    _, err = tx.Exec("UPDATE directors SET firstname = $1, lastname = $2 WHERE id = $3",
        movie.Director.Firstname, movie.Director.Lastname, movie.Director.ID)
    if err != nil {
        tx.Rollback()
        log.Printf("Error updating director: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Delete existing category associations
    _, err = tx.Exec("DELETE FROM movie_categories WHERE movie_id = $1", id)
    if err != nil {
        tx.Rollback()
        log.Printf("Error deleting existing categories: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Insert new category associations
    for _, category := range movie.Categories {
        var categoryID int
        err = tx.QueryRow("SELECT id FROM categories WHERE name = $1", category.Name).Scan(&categoryID)
        if err == sql.ErrNoRows {
            err = tx.QueryRow("INSERT INTO categories (name) VALUES ($1) RETURNING id", category.Name).Scan(&categoryID)
        }
        if err != nil {
            tx.Rollback()
            log.Printf("Error getting or creating category: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        _, err = tx.Exec("INSERT INTO movie_categories (movie_id, category_id) VALUES ($1, $2)", id, categoryID)
        if err != nil {
            tx.Rollback()
            log.Printf("Error inserting movie category: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    if err = tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = cleanupEmptyCategories(); err != nil {
        log.Printf("Error cleaning up empty categories: %v", err)
    }

    log.Println("Movie updated successfully")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(movie)
}

//Category management functions
func getCategories(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query(`
        SELECT DISTINCT c.id, c.name 
        FROM categories c
        INNER JOIN movie_categories mc ON c.id = mc.category_id
        ORDER BY c.name
    `)
    if err != nil {
        log.Printf("Error fetching categories: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var categories []Category
    for rows.Next() {
        var c Category
        if err := rows.Scan(&c.ID, &c.Name); err != nil {
            log.Printf("Error scanning category: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        categories = append(categories, c)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(categories)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
    var category Category
    err := json.NewDecoder(r.Body).Decode(&category)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err = db.QueryRow("INSERT INTO categories (name) VALUES ($1) RETURNING id", category.Name).Scan(&category.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(category)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    var category Category
    err := json.NewDecoder(r.Body).Decode(&category)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err = db.Exec("UPDATE categories SET name = $1 WHERE id = $2", category.Name, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    category.ID, _ = strconv.Atoi(id)
    json.NewEncoder(w).Encode(category)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    _, err := db.Exec("DELETE FROM categories WHERE id = $1", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Category deleted successfully"})
}

func getMoviesByCategory(w http.ResponseWriter, r *http.Request) {
    categoryID := mux.Vars(r)["id"]

    rows, err := db.Query(`
        SELECT DISTINCT m.id, m.mid, m.title, m.cover, d.id, d.firstname, d.lastname
        FROM movies m
        JOIN directors d ON m.director_id = d.id
        JOIN movie_categories mc ON m.id = mc.movie_id
        WHERE mc.category_id = $1
    `, categoryID)
    if err != nil {
        log.Printf("Error fetching movies by category: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var movies []Movie
    for rows.Next() {
        var m Movie
        err := rows.Scan(&m.ID, &m.MID, &m.Title, &m.Cover, &m.Director.ID, &m.Director.Firstname, &m.Director.Lastname)
        if err != nil {
            log.Printf("Error scanning movie: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        movies = append(movies, m)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(movies)
}

func cleanupEmptyCategories() error {
    _, err := db.Exec(`
        DELETE FROM categories
        WHERE id NOT IN (
            SELECT DISTINCT category_id
            FROM movie_categories
        )
    `)
    return err
}

func createMovie(w http.ResponseWriter, r *http.Request) {
    var movie Movie
    err := json.NewDecoder(r.Body).Decode(&movie)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	if err := validateMovie(movie); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    movie.MID = generateMID(movie)

    tx, err := db.Begin()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var directorID int
    err = tx.QueryRow("SELECT id FROM directors WHERE firstname = $1 AND lastname = $2",
        movie.Director.Firstname, movie.Director.Lastname).Scan(&directorID)
    if err == sql.ErrNoRows {
        //Director doesn't exist, create new
        err = tx.QueryRow("INSERT INTO directors (firstname, lastname) VALUES ($1, $2) RETURNING id",
            movie.Director.Firstname, movie.Director.Lastname).Scan(&directorID)
    }
    if err != nil {
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    err = tx.QueryRow("INSERT INTO movies (mid, title, director_id, cover) VALUES ($1, $2, $3, $4) RETURNING id",
        movie.MID, movie.Title, directorID, movie.Cover).Scan(&movie.ID)
    if err != nil {
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    for _, category := range movie.Categories {
        var categoryID int
        err = tx.QueryRow("SELECT id FROM categories WHERE name = $1", category.Name).Scan(&categoryID)
        if err == sql.ErrNoRows {
            err = tx.QueryRow("INSERT INTO categories (name) VALUES ($1) RETURNING id", category.Name).Scan(&categoryID)
        }
        if err != nil {
            tx.Rollback()
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        _, err = tx.Exec("INSERT INTO movie_categories (movie_id, category_id) VALUES ($1, $2)", movie.ID, categoryID)
        if err != nil {
            tx.Rollback()
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    if err = tx.Commit(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(movie)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    tx, err := db.Begin()
    if err != nil {
        log.Printf("Error beginning transaction: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //Delete associated categories first
    _, err = tx.Exec("DELETE FROM movie_categories WHERE movie_id = $1", id)
    if err != nil {
        tx.Rollback()
        log.Printf("Error deleting movie categories: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    //Delete the movie
    result, err := tx.Exec("DELETE FROM movies WHERE id = $1", id)
    if err != nil {
        tx.Rollback()
        log.Printf("Error deleting movie: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        tx.Rollback()
        log.Printf("Error getting rows affected: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if rowsAffected == 0 {
        tx.Rollback()
        http.Error(w, "Movie not found", http.StatusNotFound)
        return
    }

    //Delete the associated director if no other movies reference it
    _, err = tx.Exec("DELETE FROM directors WHERE id NOT IN (SELECT DISTINCT director_id FROM movies)")
    if err != nil {
        tx.Rollback()
        log.Printf("Error deleting orphaned directors: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = tx.Commit(); err != nil {
        log.Printf("Error committing transaction: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err = cleanupEmptyCategories(); err != nil {
        log.Printf("Error cleaning up empty categories: %v", err)
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Movie deleted successfully"})
}

func main() {
    initDB()
    defer db.Close()

    r := mux.NewRouter()

    r.HandleFunc("/health", healthCheck).Methods("GET")
    r.HandleFunc("/movies/search", searchMovies).Methods("GET")  // Moved up
    r.HandleFunc("/movies", getMovies).Methods("GET")
    r.HandleFunc("/movies", createMovie).Methods("POST")
    r.HandleFunc("/movies/{id}", getMovies).Methods("GET")
    r.HandleFunc("/movies/{id}", updateMovie).Methods("PUT")
    r.HandleFunc("/movies/{id}", deleteMovie).Methods("DELETE")
    r.HandleFunc("/categories", getCategories).Methods("GET")
    r.HandleFunc("/categories", createCategory).Methods("POST")
    r.HandleFunc("/categories/{id}", updateCategory).Methods("PUT")
    r.HandleFunc("/categories/{id}", deleteCategory).Methods("DELETE")
    r.HandleFunc("/categories/{id}/movies", getMoviesByCategory).Methods("GET")

    c := cors.New(cors.Options{
        AllowedOrigins: []string{"http://localhost:8080"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"*"},
        AllowCredentials: true,
    })

    handler := c.Handler(r)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8000"
    }
    fmt.Printf("Starting server at port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}