package main

import (
	"fmt"
	"log"
	"net/http"
	_ "regexp"
	"strings"
	_ "testing"
	"time"

	"database/sql"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	_ "github.com/aws/aws-sdk-go/service/rds/rdsutils"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	endpoint := "wildviewdb.cozovlbefpqs.us-west-2.rds.amazonaws.com:3344"
	region := "us-west-2a"
	user := "sujunzhu"
	awsCreds := credentials.NewEnvCredentials()

	// expectedRegex := `^prod-instance\.us-east-1\.rds\.amazonaws\.com:3306\?Action=connect.*?DBUser=mysqlUser.*`

	if !(strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://")) {
		endpoint = "https://" + endpoint
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		log.Fatal(err.Error())
	}
	values := req.URL.Query()
	values.Set("Action", "connect")
	values.Set("DBUser", user)
	req.URL.RawQuery = values.Encode()

	signer := v4.Signer{
		Credentials: awsCreds,
	}
	_, err = signer.Presign(req, nil, "rds-db", region, 15*time.Minute, time.Now())
	if err != nil {
    fmt.Printf("Sign Error")
		log.Fatal(err.Error())
	}

	url := req.URL.String()
	if strings.HasPrefix(url, "http://") {
		url = url[len("http://"):]
	} else if strings.HasPrefix(url, "https://") {
		url = url[len("https://"):]
	}

	//authToken, err := BuildAuthToken(endpoint, region, user, awsCreds)
	authToken := url

	// Create the MySQL DNS string for the DB connection
	// user:password@protocol(endpoint)/dbname?<params>
	dnsStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?tls=true",
		user, authToken, endpoint, "wildviewdb",
	)

	// Use db to perform SQL operations on database
	db, err := sql.Open("mysql", dnsStr)

	if err := db.Ping(); err != nil {
    fmt.Printf("Sign Error connection refused")
		log.Fatal(err.Error())
	}
}
