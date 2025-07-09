package telegrambot

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/globallstudent/academy/internal/config"
	"github.com/globallstudent/academy/internal/database"
	"gopkg.in/telebot.v3"
)

// Bot represents the Telegram bot service
type Bot struct {
	bot       *telebot.Bot
	redis     *database.Redis
	db        *database.DB
	cfg       *config.Config
	serverURL string
	loginURL  string
}

// PhoneState stores user's phone number during authentication flow
type PhoneState struct {
	PhoneNumber string
	FullName    string
	OTP         string
}

var (
	// States stores the current state for each user
	userStates = make(map[int64]*PhoneState)
)

// New creates a new Telegram bot
func New(cfg *config.Config, redis *database.Redis, db *database.DB, serverURL string) (*Bot, error) {
	// Initialize the bot with the token from config
	botToken := cfg.Telegram.BotToken
	if botToken == "" {
		return nil, fmt.Errorf("telegram bot token not set")
	}

	b, err := telebot.NewBot(telebot.Settings{
		Token:  botToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	bot := &Bot{
		bot:       b,
		redis:     redis,
		db:        db,
		cfg:       cfg,
		serverURL: serverURL,
		loginURL:  fmt.Sprintf("%s/verify", serverURL),
	}

	// Set up the bot handlers
	bot.setupHandlers()

	return bot, nil
}

// Start launches the Telegram bot
func (b *Bot) Start() {
	b.bot.Start()
}

// Stop stops the Telegram bot
func (b *Bot) Stop() {
	b.bot.Stop()
}

// setupHandlers configures all the bot's command and message handlers
func (b *Bot) setupHandlers() {
	// Command handlers
	b.bot.Handle("/start", func(c telebot.Context) error {
		return b.handleStart(c)
	})
	b.bot.Handle("/login", func(c telebot.Context) error {
		return b.handleLogin(c)
	})

	// Contact button handler
	b.bot.Handle(telebot.OnContact, func(c telebot.Context) error {
		return b.handleContact(c)
	})

	// Text message handler (for name input)
	b.bot.Handle(telebot.OnText, func(c telebot.Context) error {
		return b.handleText(c)
	})
}

// handleStart handles the /start command
func (b *Bot) handleStart(c telebot.Context) error {
	return c.Send(
		"👋 Welcome to Summer Academy Bot!\n\n" +
			"I will help you authenticate for the Summer Academy platform.\n\n" +
			"Use /login to start the authentication process.",
	)
}

// handleLogin handles the /login command
func (b *Bot) handleLogin(c telebot.Context) error {
	userID := c.Sender().ID

	// Reset user state if it exists
	userStates[userID] = &PhoneState{}

	// Create a custom keyboard for phone number sharing
	kb := &telebot.ReplyMarkup{ResizeKeyboard: true}
	shareBtn := kb.Contact("📱 Share Phone Number")
	kb.Reply(kb.Row(shareBtn))

	return c.Send(
		"Please share your phone number to authenticate.\n\n"+
			"Your number will be securely stored and used only for authentication.",
		kb,
	)
}

// handleContact handles phone number sharing
func (b *Bot) handleContact(c telebot.Context) error {
	userID := c.Sender().ID
	contact := c.Message().Contact

	if contact == nil {
		return c.Send("Please use the 'Share Phone Number' button.")
	}

	// Ensure this contact belongs to the user
	if contact.UserID != userID {
		return c.Send("Please share your own contact.")
	}

	// Format the phone number
	phoneNumber := contact.PhoneNumber
	if phoneNumber == "" {
		return c.Send("Could not get your phone number. Please try again with /login.")
	}

	// Create or update user state
	state := userStates[userID]
	if state == nil {
		state = &PhoneState{}
		userStates[userID] = state
	}
	state.PhoneNumber = phoneNumber

	// Remove custom keyboard
	kb := &telebot.ReplyMarkup{RemoveKeyboard: true}

	// Ask for full name
	return c.Send(
		"Great! Now, please enter your full name:",
		kb,
	)
}

// handleText handles text messages (for name input)
func (b *Bot) handleText(c telebot.Context) error {
	userID := c.Sender().ID
	state := userStates[userID]

	// If no active state or phone number, ask to start login process
	if state == nil || state.PhoneNumber == "" {
		return c.Send("Please use /login to start the authentication process.")
	}

	// If name is not set yet, this is the name input
	if state.FullName == "" {
		fullName := c.Text()
		if len(fullName) < 3 {
			return c.Send("Please enter your real full name (at least 3 characters).")
		}

		// Save full name
		state.FullName = fullName

		// Generate OTP
		otp := generateOTP()
		state.OTP = otp

		// Store OTP in Redis with expiration (5 minutes) if Redis is available
		if b.redis != nil {
			err := b.redis.StoreOTP(state.PhoneNumber, otp, 5*time.Minute)
			if err != nil {
				// Log the error in production
				log.Printf("Error storing OTP in Redis: %v", err)
				return c.Send("An error occurred while storing OTP. Please try again later.")
			}
		} else {
			// In development mode without Redis, just log the OTP
			log.Printf("Development mode: OTP for %s is %s", state.PhoneNumber, otp)
		}

		// Create login URL with parameters
		loginURL := fmt.Sprintf("%s?phone=%s&otp=%s", b.loginURL, state.PhoneNumber, otp)

		// Send OTP and login link
		return c.Send(
			fmt.Sprintf("Your verification code is: *%s*\n\n", otp)+
				"You can use this code on the website, or simply click the link below:\n\n"+
				fmt.Sprintf("[Click here to login](%s)", loginURL),
			&telebot.SendOptions{
				ParseMode: telebot.ModeMarkdown,
			},
		)
	}

	// If we reach here, user has already received their OTP but sent another message
	return c.Send(
		"You already have an active login request.\n" +
			"Please use the login link I sent you, or use /login to start again.",
	)
}

// generateOTP generates a 6-digit random OTP
func generateOTP() string {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number between 100000 and 999999
	otp := 100000 + rand.Intn(900000)

	return fmt.Sprintf("%06d", otp)
}
