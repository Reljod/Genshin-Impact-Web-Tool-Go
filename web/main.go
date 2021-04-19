package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	_ "github.com/lib/pq"
)

var pgDb *sql.DB

func main() {

	var err error
	pgDb, err = CreateDatabase()
	CheckError(err)

	defer pgDb.Close()

	router := mux.NewRouter()
	router.HandleFunc("/", homePage).Methods("GET")
	router.HandleFunc("/characters", characterList).Methods("GET")
	router.HandleFunc("/characters/add", postAddCharacter).Methods("POST")

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})

	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %s...\n", addr)
	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func CreateDatabase() (*sql.DB, error) {
	psqlconn := os.Getenv("DATABASE_URL")

	if psqlconn == "" {
		log.Fatalf("No DATABASE_URL environment variable")
	}

	db, err := sql.Open("postgres", psqlconn)
	if err != nil {
		return nil, err
	}

	fmt.Println("Pinging PostgreSQL Database..")
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully Communicated PostgreSQL Database..")
	return db, nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../templates/index.html")
	if err != nil {
		log.Panic("Cannot parse Index.html template", err)
	}
	t.Execute(w, r)
}

func characterList(w http.ResponseWriter, r *http.Request) {

	var characterList CharacterList

	rows, err := pgDb.Query(`SELECT "id", "Name", "Vision", "Affiliation", "Gender", "Weapon Type" FROM dbo."Characters" c `)
	CheckError(err)

	defer rows.Close()
	for rows.Next() {
		var character Character
		err = rows.Scan(
			&character.Id,
			&character.Name,
			&character.Vision,
			&character.Affiliation,
			&character.Gender,
			&character.WeaponType,
		)
		CheckError(err)

		fmt.Printf("Reading %+v\n", characterList)
		characterList.addCharacter(character)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(characterList)
}

func postAddCharacter(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "This endpoint only supports POST requests", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	var character CharacterInput

	err = json.Unmarshal(body, &character)
	if err != nil {
		http.Error(w, "Cannot Parse Request Body", http.StatusBadRequest)
		return
	}

	fmt.Printf("%+v", character)

	insertDynStmt := `INSERT INTO dbo."Characters"("Name", "Vision", "Affiliation", "Gender", "Weapon Type") VALUES ($1, $2, $3, $4, $5)`
	fmt.Println(insertDynStmt)
	_, err = pgDb.Exec(insertDynStmt, character.Name, character.Vision, character.Affiliation, character.Gender, character.WeaponType)
	CheckError(err)

	w.Header().Set("Content-Type", "application/json")
	setupResponse(&w)

	jsObj, _ := json.Marshal(character)
	w.Write(jsObj)
}

func setupResponse(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

type Character struct {
	Id          string
	Name        string
	Vision      string
	Affiliation string
	Gender      string
	WeaponType  string
}

type CharacterInput struct {
	Affiliation string
	Gender      string
	Name        string
	Vision      string
	WeaponType  string
}

type CharacterList struct {
	Characters []Character
}

func (characterList *CharacterList) addCharacter(c Character) {
	characterList.Characters = append(characterList.Characters, c)
}
