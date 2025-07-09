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
	bot         *telebot.Bot
	redis       *database.Redis
	db          *database.DB
	cfg         *config.Config
	serverURL   string
	loginURL    string
	devOTPStore interface{} // Function to store development OTPs
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

	// Create the bot instance
	bot := &Bot{
		bot:         b,
		redis:       redis,
		db:          db,
		cfg:         cfg,
		serverURL:   serverURL,
		loginURL:    fmt.Sprintf("%s/verify", serverURL),
		devOTPStore: nil, // Will be set by the caller if needed
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

	// Check if we already have this user's phone number
	state, exists := userStates[userID]

	// If user hasn't shared their phone number yet
	if !exists || state == nil || state.PhoneNumber == "" {
		// Create a custom keyboard for phone number sharing
		kb := &telebot.ReplyMarkup{ResizeKeyboard: true}
		shareBtn := kb.Contact("📱 Share Phone Number")
		kb.Reply(kb.Row(shareBtn))

		return c.Send(
			"Please share your phone number to authenticate.\n\n"+
				"You'll only need to do this once.",
			kb,
		)
	}

	// User already has shared their phone number before, generate a new OTP
	// Generate OTP
	otp := generateOTP()
	state.OTP = otp

	// Store OTP in Redis with expiration (5 minutes) if Redis is available
	if b.redis != nil && b.redis.Client != nil {
		err := b.redis.StoreOTP(state.PhoneNumber, otp, 5*time.Minute)
		if err != nil {
			// Log the error in production
			log.Printf("Error storing OTP in Redis: %v", err)
			return c.Send("An error occurred while storing OTP. Please try again later.")
		}
	} else {
		// In development mode without Redis
		if storeFunc, ok := b.devOTPStore.(func(string, string)); ok {
			storeFunc(state.PhoneNumber, otp)
		}
		log.Printf("Development mode: OTP for %s is %s", state.PhoneNumber, otp)
	}

	// Send verification message with OTP and quick login links
	return b.sendVerificationMessage(c, state.PhoneNumber, otp)
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
	c.Send("Thank you!", kb)

	// Generate OTP immediately
	otp := generateOTP()
	state.OTP = otp

	// Store the username if available, otherwise use a default name
	username := c.Sender().Username
	if username == "" {
		username = "User"
	}
	state.FullName = username

	// Store OTP in Redis with expiration (5 minutes) if Redis is available
	if b.redis != nil && b.redis.Client != nil {
		err := b.redis.StoreOTP(state.PhoneNumber, otp, 5*time.Minute)
		if err != nil {
			// Log the error in production
			log.Printf("Error storing OTP in Redis: %v", err)
			return c.Send("An error occurred while storing OTP. Please try again later.")
		}
	} else {
		// In development mode without Redis
		if storeFunc, ok := b.devOTPStore.(func(string, string)); ok {
			storeFunc(state.PhoneNumber, otp)
		}
		log.Printf("Development mode: OTP for %s is %s", state.PhoneNumber, otp)
	}

	// Send verification message with OTP and quick login links
	return b.sendVerificationMessage(c, state.PhoneNumber, otp)
}

// handleText handles text messages
func (b *Bot) handleText(c telebot.Context) error {
	text := c.Text()

	// If the text is /login, handle it as the login command
	if text == "/login" {
		return b.handleLogin(c)
	}

	userID := c.Sender().ID
	state := userStates[userID]

	// If no active state or phone number, ask to start login process
	if state == nil || state.PhoneNumber == "" {
		return c.Send("Please use /login to start the authentication process.")
	}

	// If user already has an active session, provide help
	return c.Send("Need a new login code? Use /login to get one.\n\nIf you're having issues, try /start to restart the bot.")
}

// SetDevOTPStore sets the function to use for development mode OTP storage
func (b *Bot) SetDevOTPStore(storeFunc interface{}) {
	b.devOTPStore = storeFunc
}

// generateOTP generates a 6-digit random OTP
func generateOTP() string {
	// Seed the random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random number between 100000 and 999999
	otp := 100000 + rand.Intn(900000)

	return fmt.Sprintf("%06d", otp)
}
