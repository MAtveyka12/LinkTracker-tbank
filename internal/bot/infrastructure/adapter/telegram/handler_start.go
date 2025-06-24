package telegram

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"text/template"

	"gopkg.in/telebot.v3"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
)

func formatUserDetails(user domain.User) (string, error) {
	templatePath := os.Getenv("TMPL_USER_DETAILS_FILE_PATH")
	if templatePath == "" {
		return "", fmt.Errorf("TMPL_USER_DETAILS_FILE_PATH environment variable not set")
	}

	tmplContent, err := os.ReadFile(templatePath)

	if err != nil {
		return "", fmt.Errorf("error reading the template %s: %s", templatePath, err.Error())
	}

	tmpl, err := template.New("userDetails").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("error when parsing the template: %s", err.Error())
	}

	var message bytes.Buffer
	if err := tmpl.Execute(&message, user); err != nil {
		return "", fmt.Errorf("error in message generation: %s", err.Error())
	}

	return message.String(), nil
}

func (t *TgBot) Start(ctx telebot.Context) error {
	userID := ctx.Sender().ID

	logger := slog.With(
		slog.Int64("user_id", userID),
		slog.String("handler", "start"),
	)

	if user, exists := domain.Users[userID]; exists {
		message, err := formatUserDetails(user)
		if err != nil {
			slog.Error(fmt.Sprintf("Error formatting user data: %s", err.Error()))
			return ctx.Send("Ошибка при получении ваших данных. Попробуйте позже.")
		}

		return ctx.Send(message)
	}

	err := ctx.Send("Введите ваш email:")

	if err != nil {
		return fmt.Errorf("incorrect data: %s", err.Error())
	}

	domain.UserRegistrationData[userID] = domain.UsersRegistrationData{
		State: domain.StateWaitingForEmail,
		Email: "",
	}

	logger.Info("Started registration process",
		slog.Int("current_state", domain.StateWaitingForEmail),
	)

	return nil
}

func (t *TgBot) HandleRegistrationEmail(ctx telebot.Context, userID int64, data domain.UsersRegistrationData) error {
	logger := slog.With(
		slog.Int64("user_id", userID),
		slog.String("handler", "registration_email"),
	)

	email := ctx.Text()
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return ctx.Send("Некорректный формат email. Пожалуйста, введите email в формате example@example.com.")
	}

	data.Email = email
	data.State = domain.StateWaitingForPassword
	domain.UserRegistrationData[userID] = data

	logger.Info("Email accepted, requesting password",
		slog.Int("new_state", domain.StateWaitingForPassword),
	)

	return ctx.Send("Введите ваш пароль:")
}

func (t *TgBot) HandleRegistrationPassword(ctx telebot.Context, userID int64, data domain.UsersRegistrationData) error {
	logger := slog.With(
		slog.Int64("user_id", userID),
		slog.String("handler", "registration_password"),
	)

	password := ctx.Text()
	newUser := domain.User{
		TelegramID: userID,
		UserName:   ctx.Sender().Username,
		Email:      data.Email,
		Password:   password,
	}

	logger.Info("Registering new user",
		slog.String("email", newUser.Email),
		slog.String("username", newUser.UserName),
	)

	err := t.Scrapper.RegisterUser(context.Background(), newUser)
	if err != nil {
		logger.Error("Failed to register user with scrapper",
			slog.String("error", err.Error()),
		)

		if err := ctx.Send("Произошла ошибка при регистрации. Пожалуйста, попробуйте снова."); err != nil {
			slog.Error(fmt.Sprintf("couldn't send message: %s", err.Error()))
		}

		return fmt.Errorf("error when registering a user for a scrapper: %s", err.Error())
	}

	domain.Users[userID] = newUser

	delete(domain.UserRegistrationData, userID)

	logger.Info("User successfully registered")

	return ctx.Send("Вы успешно зарегистрировались.")
}

func (t *TgBot) StartHandler() {
	slog.Info("Registering Telegram command handlers")

	t.Bot.Handle("/start", t.Start)
}
