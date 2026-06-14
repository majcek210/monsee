package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/term"

	"github.com/majcek210/monsee/internal/repository/postgres"
	"github.com/majcek210/monsee/internal/service"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepo(pool)
	userSvc := service.NewUserService(userRepo)

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Password: ")
	pwBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		log.Fatalf("read password: %v", err)
	}
	password := strings.TrimSpace(string(pwBytes))

	u, err := userSvc.Register(context.Background(), email, password, "admin")
	if err != nil {
		log.Fatalf("create admin: %v", err)
	}

	fmt.Printf("Admin created: %s (%s)\n", u.Email, u.ID)
}
