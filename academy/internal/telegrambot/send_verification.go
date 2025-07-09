package telegrambot

import (
	"fmt"
	"net/url"

	"gopkg.in/telebot.v3"
)

// sendVerificationMessage sends an OTP code and verification link to the user
func (b *Bot) sendVerificationMessage(c telebot.Context, phoneNumber, otp string) error {
	// Create verification URLs
	verifyURL := fmt.Sprintf("%s?phone=%s&otp=%s",
		b.loginURL, url.QueryEscape(phoneNumber), otp)

	// Create a pre-filled direct login URL
	loginURL := fmt.Sprintf("%s/login?phone=%s&otp=%s",
		b.serverURL, url.QueryEscape(phoneNumber), otp)

	// Format a message with markdown
	message := fmt.Sprintf("🔐 *Verification Code*\n\n"+
		"Your login code is: *%s*\n\n"+
		"📱 For phone number: *%s*\n\n"+
		"🔗 *Quick Login Options*:\n"+
		"1️⃣ [Click here to verify automatically](%s)\n"+
		"2️⃣ [Open the login page](%s)\n\n"+
		"Or enter the code manually on the login page.",
		otp, phoneNumber, verifyURL, loginURL)

	// Send the message with markdown formatting
	return c.Send(message, &telebot.SendOptions{
		ParseMode: telebot.ModeMarkdown,
	})
}
