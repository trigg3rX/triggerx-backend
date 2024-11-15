package database

import (
	"fmt"
	"log"
	"time"
	"github.com/gocql/gocql"
)

func TestDatabaseConnection() {
	fmt.Println("Testing Database Connection...")
	
	// Initialize database
	config := NewConfig()
	conn, err := NewConnection(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Initialize schema
	if err := InitSchema(conn.Session()); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Test inserting data
	if err := testInsertData(conn.Session()); err != nil {
		log.Fatalf("Failed to insert test data: %v", err)
	}

	// Test retrieving data
	if err := testRetrieveData(conn.Session()); err != nil {
		log.Fatalf("Failed to retrieve test data: %v", err)
	}

	fmt.Println("Database tests completed successfully!")
}

func testInsertData(session *gocql.Session) error {
	// Generate a test UUID
	id := gocql.TimeUUID()
	
	// Insert test data with keyspace prefix
	if err := session.Query(`
		INSERT INTO triggerx.tasks (id, name, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		id, "Test Task", "pending", time.Now(), time.Now(),
	).Exec(); err != nil {
		return fmt.Errorf("failed to insert test data: %v", err)
	}

	fmt.Println("Test data inserted successfully")
	return nil
}

func testRetrieveData(session *gocql.Session) error {
	// Query the tasks table with keyspace prefix
	iter := session.Query(`SELECT id, name, status FROM triggerx.tasks LIMIT 1`).Iter()
	
	var id gocql.UUID
	var name, status string

	if iter.Scan(&id, &name, &status) {
		fmt.Printf("Retrieved task - ID: %s, Name: %s, Status: %s\n", id, name, status)
	} else {
		return fmt.Errorf("no data found in tasks table")
	}

	if err := iter.Close(); err != nil {
		return fmt.Errorf("error closing iterator: %v", err)
	}

	return nil
} 