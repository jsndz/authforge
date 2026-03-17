package email

type EmailService struct {
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (s *EmailService) SendEmailVerification(email, token string) error {
	// Implementation for sending email verification
	return nil
}
