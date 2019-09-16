package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"net/http"

	"github.com/gin-gonic/gin"
)

type Customer struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var customers []Customer
var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Println("connect to database error", err)
	}
	defer db.Close()

	createTable := `
	CREATE TABLE IF  NOT EXISTS customers (
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT,
		status TEXT
	);
		`
	_, err = db.Exec(createTable)
	if err != nil {
		log.Println("connect to database error", err)
	}

	r := gin.Default()
	r.Use(authMiddleware)
	r.POST("/customers", insertCustHandler)
	r.GET("/customers/:id", getLatestCustHandler)
	r.GET("/customers", getAllCustHandler)
	r.PUT("/customers/:id", updateLatestCustHandler)
	r.DELETE("/customers/:id", deleteLatestCustHandler)
	//
	r.Run(":2019") // default port 8080
}

func getAllCustHandler(c *gin.Context) {
	customers = nil

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect fail", err)
	}

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers")
	if err != nil {
		log.Println("can't prepare query all customers statment", err)
	}
	rows, err := stmt.Query()
	if err != nil {
		log.Println("can't query all customers", err)
	}
	for rows.Next() {
		cust := Customer{}
		err := rows.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
		if err != nil {
			log.Println("can't Scan row into variable", err)
		}

		customers = append(customers, cust)
	}
	defer db.Close()
	c.JSON(http.StatusOK, customers)

}

func getLatestCustHandler(c *gin.Context) {
	var cust Customer
	err := c.ShouldBindJSON(&cust)

	id := c.Param("id")
	idnum, err := strconv.Atoi(id)
	cust.ID = idnum

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect fail", err)
	}

	stmt, err := db.Prepare("SELECT id, name, email, status FROM customers where id=$1")
	if err != nil {
		log.Println("can'tprepare query one row statment", err)
	}

	row := stmt.QueryRow(cust.ID)
	err = row.Scan(&cust.ID, &cust.Name, &cust.Email, &cust.Status)
	if err != nil {
		log.Println("can't Scan row into variables", err)
	}
	defer db.Close()
	c.JSON(http.StatusOK, cust)

}

func insertCustHandler(c *gin.Context) {
	var cust Customer
	err := c.ShouldBindJSON(&cust)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect fail", err)
	}

	row := db.QueryRow("INSERT INTO customers (name , email , status) values ($1, $2, $3) RETURNING id", cust.Name, cust.Email, cust.Status)
	var id int
	err = row.Scan(&id)
	if err != nil {
		fmt.Println("can't scan id", err)
		return
	}
	// fmt.Println("insert customers success id : ", id)
	defer db.Close()
	cust.ID = id

	c.JSON(http.StatusCreated, cust)

}

func updateLatestCustHandler(c *gin.Context) {
	var cust Customer
	err := c.ShouldBindJSON(&cust)

	id := c.Param("id")
	idnum, err := strconv.Atoi(id)
	cust.ID = idnum

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect fail", err)
	}

	stmt, err := db.Prepare("UPDATE customers SET name=$2 , email=$3 , status=$4 WHERE id=$1;")
	if err != nil {
		log.Println("can't prepare statment update", err)
	}
	if _, err := stmt.Exec(cust.ID, cust.Name, cust.Email, cust.Status); err != nil {
		log.Println("error execute update ", err)
	}
	// fmt.Println("update success")
	c.JSON(http.StatusOK, cust)

}

func deleteLatestCustHandler(c *gin.Context) {
	var cust Customer
	err := c.ShouldBindJSON(&cust)

	id := c.Param("id")
	idnum, err := strconv.Atoi(id)
	cust.ID = idnum

	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println("Connect fail", err)
	}

	stmt, err := db.Prepare("DELETE FROM customers WHERE id=$1;")
	if err != nil {
		log.Println("can't prepare statment update", err)
	}
	if _, err := stmt.Exec(cust.ID); err != nil {
		log.Println("error execute update ", err)
	}
	// fmt.Println("delete success")
	c.JSON(http.StatusOK, gin.H{"message": "customers deleted"})

}

func authMiddleware(c *gin.Context) {
	fmt.Println("This is a middlewear")
	token := c.GetHeader("Authorization")

	if token != "token2019" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return
	}
	c.Next()

}
