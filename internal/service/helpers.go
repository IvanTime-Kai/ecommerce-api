package service

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"golang.org/x/crypto/bcrypt"
)

func toNullString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func generateSlug(name string) string {
	lowerName := strings.ToLower(name)
	slug := strings.ReplaceAll(lowerName, " ", "-")
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	return fmt.Sprintf("%s-%s", slug, uuid.New().String()[:8])
}
