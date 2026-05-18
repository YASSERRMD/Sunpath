package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	st, err := store.NewPostgresStore(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer st.Close()

	fmt.Println("migrations complete")
}
