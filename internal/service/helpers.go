package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ivantime-Kai/ecommerce-api/internal/repository"
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

func numericToFloat(n pgtype.Numeric) float64 {
	f, _ := n.Float64Value()
	return f.Float64
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric
	n.Scan(strconv.FormatFloat(f, 'f', 2, 64))
	return n
}

func isValidTransition(current, next repository.OrderStatus) bool {
	allowed := map[repository.OrderStatus][]repository.OrderStatus{
		repository.OrderStatusPending:   {repository.OrderStatusConfirmed, repository.OrderStatusCancelled},
		repository.OrderStatusConfirmed: {repository.OrderStatusShipping, repository.OrderStatusCancelled},
		repository.OrderStatusShipping:  {repository.OrderStatusDelivered},
		repository.OrderStatusDelivered: {},
		repository.OrderStatusCancelled: {},
	}
	for _, s := range allowed[current] {
		if s == next {
			return true
		}
	}
	return false
}
