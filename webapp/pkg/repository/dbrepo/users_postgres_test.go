//go:build integration

// (to run the tests --> go test -v -tags=<tag-name> . --> eg go test -v -tags=integration .)
package dbrepo

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
	"webapp/pkg/data"
	"webapp/pkg/repository"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "postgres"
	password = "postgres"
	dbName   = "users_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var resource *dockertest.Resource
var pool *dockertest.Pool
var testDB *sql.DB
var testRepo repository.DatabaseRepo

func TestMain(m *testing.M) {
	// connect to docker; fail if docker not running
	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker; is it running ?? %s", err)
	}

	pool = p

	// setup up our docker options, specifying the image and so forth
	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14.5", // image tag
		Env: []string{ // env variables that we want to use
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"}, //exposed ports in the container --> internal port inside the container that is exposed
		PortBindings: map[docker.Port][]docker.PortBinding{ //ports exposed on my local machine
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}
	// get a resource (an instance of the docker image)
	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource: %s", err)
	}

	// start the image and wait until the docker image is ready
	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			log.Println("error while trying to connect to pg instance running on docker: ", err)
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to pg instance running on docker. Error: %s", err)
	}

	// populate the database with empty tables
	err = createTables()
	if err != nil {
		log.Fatalf("error creating tables in pg instance running on docker: %s", err)
	}

	// setup a database connection pool
	testRepo = &PostgresDBRepo{DB: testDB}

	// run the tests
	code := m.Run()

	// cleanup --> stop everything and purge the resource after the tests have completed
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge the resource: %s", err)
	}

	os.Exit(code)
}

// createTables --> populates the pg instance running on docker with empty tables
func createTables() error {
	tableSQL, err := os.ReadFile("./testdata/users.sql")
	if err != nil {
		fmt.Printf("error while reading sql file in testdata folder: %s", err)
		return err
	}

	_, err = testDB.Exec(string(tableSQL))
	if err != nil {
		fmt.Printf("error while executing SQL file in testdata folder: %s", err)
		return err
	}

	return nil
}

func Test_pingDB(t *testing.T) {
	err := testDB.Ping()
	if err != nil {
		t.Error("can not ping test pg instance running on docker")
	}
}

// Test inserting a user
func TestPostgresDBRepoInsertUser(t *testing.T) {
	testUser := data.User{
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := testRepo.InsertUser(testUser)
	if err != nil {
		t.Errorf("insert user returned an error: %s", err)
	}
	if id != 1 {
		t.Errorf("insert user returned wrong id. expected %d, but got %d", 1, id)
	}
}

// Test get all users
func TestPostgresDBRepoGetAllUsers(t *testing.T) {
	users, err := testRepo.AllUsers()
	if err != nil {
		t.Errorf("all users reports an error: %s", err)
	}
	if len(users) != 1 {
		t.Errorf("all users reports wrong size; expected %d, but got %d", 1, len(users))
	}

	// insert another user and check the length of the result
	testUser := data.User{
		FirstName: "Jack",
		LastName:  "Smith",
		Email:     "jack@smith.com",
		Password:  "secret",
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, _ = testRepo.InsertUser(testUser)

	users, err = testRepo.AllUsers()
	if err != nil {
		t.Errorf("get all users reports an error: %s", err)
	}
	if len(users) != 2 {
		t.Errorf("all users reports wrong size after insert; expected %d, but got %d", 2, len(users))
	}

	// TODO: check if all users are sorted alphabetically --> eg when sorting by first name
	// TODO: check to ensure that you can't insert users with the same email
}

// Getting individual users --> eg by email or ID
func TestPostgresDBRepoGetUser(t *testing.T) {
	user, err := testRepo.GetUser(1)
	if err != nil {
		t.Errorf("error getting user by ID: %s", err)
	}
	if user.Email != "admin@example.com" {
		t.Errorf("wrong user returned by GetUser; expected %s, but got %s", "admin@example.com", user.Email)
	}

	// check for a user that doesn't exist
	_, err = testRepo.GetUser(100)
	if err == nil {
		t.Errorf("no error reported when getting non existent user by ID: %d", 100)
	}
	// TODO: check for non-existent user ID eg 100
}

// Get individual users --> by email
func TestPostgresDBRepoGetUserByEmail(t *testing.T) {
	user, err := testRepo.GetUserByEmail("jack@smith.com")
	if err != nil {
		t.Errorf("error getting user by email: %s", err)
	}
	if user.ID != 2 {
		t.Errorf("wrong ID returned by GetUserByEmail; expected %d, but got %d", 2, user.ID)
	}
	// TODO: check for a non-existent email
}

// Update a user
func TestPostgresDBRepoUpdateUser(t *testing.T) {
	user, _ := testRepo.GetUser(2)
	user.FirstName = "Jane"
	user.Email = "jane@smith.com"

	err := testRepo.UpdateUser(*user)
	if err != nil {
		t.Errorf("error while updating user with id %d: %s", 2, err)
	}

	user, _ = testRepo.GetUser(2)
	if user.FirstName != "Jane" || user.Email != "jane@smith.com" {
		t.Errorf("error while updating user details. expected first name to be %s, but got %s. expected email to be %s but got %s", "Jane", user.FirstName, "janes@smith.com", user.Email)
	}
}

// delete user
func TestPostgresDBRepoDeleteUser(t *testing.T) {
	err := testRepo.DeleteUser(2)
	if err != nil {
		t.Errorf("error while deleting user with id %d: %s", 2, err)
	}

	// try to get the deleted user
	_, err = testRepo.GetUser(2)
	if err == nil {
		t.Errorf("expected an error while retrieving a user that has already been deleted but didn't get any")
	}

	users, _ := testRepo.AllUsers()
	if len(users) != 1 {
		t.Errorf("error while deleting a user. expected length to be %d, but got %d", 1, len(users))
	}
}

// reset password
func TestPostgresDBRepoResetPassword(t *testing.T) {
	err := testRepo.ResetPassword(1, "password") // use id=1, because we have already deleted the second user in the prior test
	if err != nil {
		t.Errorf("error resetting password for user: %d", 2)
	}
	// check to ensure that the password has been updated
	user, _ := testRepo.GetUser(1)
	matches, err := user.PasswordMatches("password")
	if err != nil {
		t.Error(err)
	}
	if !matches {
		t.Errorf("password should match 'password', but does not")
	}
}

// insert user image
func TestPostgresDBRepoInsertUserImage(t *testing.T) {
	var image data.UserImage
	image.UserID = 1
	image.FileName = "test.jpg"
	image.CreatedAt = time.Now()
	image.UpdatedAt = time.Now()

	newID, err := testRepo.InsertUserImage(image)
	if err != nil {
		t.Errorf("error while inserting user image: %s", err)
	}
	if newID != 1 { // since it's the first insertion --> first record
		t.Errorf("expected user image ID to be %d, but got %d: ", 1, newID)
	}

	// assign image USERID to a user that doesn't exists in the DB
	image.UserID = 100
	_, err = testRepo.InsertUserImage(image)
	if err == nil {
		t.Errorf("expected error for a user id: %v which doesn't exists, but didn't get any", image.UserID)
	}

	//TODO: refactor this and other tests to table driven tests
}
