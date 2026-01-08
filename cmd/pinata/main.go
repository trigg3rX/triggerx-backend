package main

// This script is used to list all files in a Pinata account and delete files older than 12 hours with "proof_of_task" in the name

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joho/godotenv"
	"github.com/trigg3rX/triggerx-backend/pkg/env"
	"github.com/trigg3rX/triggerx-backend/pkg/ipfs"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, just log a warning
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get configuration from environment variables
	pinataHost := env.GetEnvString("PINATA_HOST", "")
	pinataJWT := env.GetEnvString("PINATA_JWT", "")

	// Validate required environment variables
	if pinataHost == "" {
		log.Fatal("Error: PINATA_HOST environment variable is required")
	}
	if pinataJWT == "" {
		log.Fatal("Error: PINATA_JWT environment variable is required")
	}

	// Create IPFS client configuration
	ipfsCfg := ipfs.NewConfig(pinataHost, pinataJWT)

	// Initialize IPFS client
	ipfsClient, err := ipfs.NewClient(ipfsCfg)
	if err != nil {
		log.Fatalf("Error: Failed to initialize IPFS client: %v", err)
	}
	defer ipfsClient.Close()

	// Create context
	ctx := context.Background()

	// List all files from Pinata
	fmt.Println("Fetching files from Pinata...")
	files, err := ipfsClient.ListFiles(ctx)
	if err != nil {
		log.Fatalf("Error: Failed to list IPFS files: %v", err)
	}

	// Display results
	fmt.Printf("\nFound %d file(s) in Pinata account\n\n", len(files))

	if len(files) == 0 {
		fmt.Println("No files found in Pinata account.")
		return
	}

	// Filter files for deletion: older than 48 hours and containing "proof_of_task" in name
	const deletionAge = 12 * time.Hour
	cutoffTime := time.Now().Add(-deletionAge)
	var filesToDelete []ipfs.PinataFile

	for _, file := range files {
		// Check if file name contains "proof_of_task" (case-insensitive)
		if strings.Contains(strings.ToLower(file.Name), "proof_of_task") {
			// Check if file is older than 48 hours
			if file.CreatedAt.Before(cutoffTime) {
				filesToDelete = append(filesToDelete, file)
			}
		}
	}

	// Display all files
	fmt.Println("All files:")
	fmt.Println("==========")
	createTabWriter(files)

	// Print summary
	fmt.Printf("\nTotal files: %d\n", len(files))

	// Delete matching files
	if len(filesToDelete) == 0 {
		fmt.Println("\nNo files found matching deletion criteria (older than 48 hours with 'proof_of_task' in name).")
		return
	}

	fmt.Printf("\nFound %d file(s) matching deletion criteria:\n", len(filesToDelete))
	fmt.Println("==============================================")
	createTabWriter(filesToDelete)

	// Ask for confirmation (or use a flag in the future)
	fmt.Printf("\nDeleting %d file(s)...\n\n", len(filesToDelete))

	var deletedCount, failedCount int
	for i, file := range filesToDelete {
		fmt.Printf("[%d/%d] Deleting file: %s (CID: %s, Created: %s)... ",
			i+1, len(filesToDelete), file.Name, file.CID, file.CreatedAt.Format(time.RFC3339))

		if err := ipfsClient.Delete(ctx, file.CID); err != nil {
			fmt.Printf("FAILED: %v\n", err)
			failedCount++
		} else {
			fmt.Println("SUCCESS")
			deletedCount++
		}
	}

	fmt.Printf("\nDeletion complete:\n")
	fmt.Printf("  Successfully deleted: %d\n", deletedCount)
	fmt.Printf("  Failed: %d\n", failedCount)
}

// createTabWriter creates and displays a formatted table of files
func createTabWriter(files []ipfs.PinataFile) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tCID\tName\tSize\tMIME Type\tCreated At"); err != nil {
		log.Printf("Error writing table header: %v", err)
		return
	}
	if _, err := fmt.Fprintln(w, "---\t---\t---\t---\t---\t---"); err != nil {
		log.Printf("Error writing table separator: %v", err)
		return
	}

	for _, file := range files {
		sizeStr := formatSize(file.Size)
		createdAtStr := file.CreatedAt.Format(time.RFC3339)

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			file.ID,
			file.CID,
			file.Name,
			sizeStr,
			file.MimeType,
			createdAtStr,
		); err != nil {
			log.Printf("Error writing file data: %v", err)
			return
		}
	}

	if err := w.Flush(); err != nil {
		log.Printf("Error flushing tab writer: %v", err)
	}
}

// formatSize formats file size in human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
