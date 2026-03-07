package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/Jiruu246/rms/internal/ent"
)

// SeedData contains all the seed data for the database
type SeedData struct {
	Users []UserSeed
}

type UserSeed struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}

// GetSeedData returns predefined seed data
func GetSeedData() SeedData {
	// User IDs
	user1ID := uuid.New()
	user2ID := uuid.New()

	return SeedData{
		Users: []UserSeed{
			{
				ID:       user1ID,
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			{
				ID:       user2ID,
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Password: "password123",
			},
		},
	}
}

// seedDatabase populates the database with initial data
func seedDatabase(ctx context.Context, client *ent.Client) error {
	log.Println("🌱 Starting database seeding...")

	seedData := GetSeedData()

	// Seed Users
	log.Println("👤 Creating users...")
	for _, userSeed := range seedData.Users {
		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userSeed.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", userSeed.Email, err)
		}

		user, err := client.User.Create().
			SetID(userSeed.ID).
			SetName(userSeed.Name).
			SetEmail(userSeed.Email).
			SetIsActive(true).
			SetPasswordHash(string(hashedPassword)).
			Save(ctx)

		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", userSeed.Email, err)
		}
		log.Printf("  ✅ Created user: %s (%s)", user.Name, user.Email)
	}

	log.Println("🎉 Database seeding completed successfully!")
	return nil
}
