package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	database = "genshin_impact_db"
)

var password string = "admin"

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

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":5000", handler))

	// http.HandleFunc("/", homePage)
	// http.HandleFunc("/characters", characterList)
	// http.HandleFunc("/characters/add", postAddCharacter)
	// log.Fatal(http.ListenAndServe(":5000", nil))
}

func CreateDatabase() (*sql.DB, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, database)
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