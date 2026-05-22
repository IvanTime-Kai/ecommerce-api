package service

import (
	"fmt"
	"net/smtp"
	"strconv"

	"github.com/Ivantime-Kai/ecommerce-api/internal/config"
	"github.com/Ivantime-Kai/ecommerce-api/internal/kafka"
)

type NotificationService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
}

func NewNotificationService(cfg *config.SMTPConfig) *NotificationService {
	port, _ := strconv.Atoi(cfg.Port)

	return &NotificationService{
		smtpHost:     cfg.Host,
		smtpPort:     port,
		smtpUsername: cfg.Username,
		smtpPassword: cfg.Password,
	}
}

func (s *NotificationService) SendOrderConfirmation(event kafka.PaymentCompletedEvent) error {
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	subject := "Subject: Order Confirmation\r\n"
	body := fmt.Sprintf("Your order %s has been placed. Total: %.0f VND", event.OrderID, event.TotalAmount)
	msg := []byte(subject + "\r\n" + body)

	return smtp.SendMail(addr, auth, s.smtpUsername, []string{event.BuyerEmail}, msg)
}
