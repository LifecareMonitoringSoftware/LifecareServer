package validate

import (
	"net/mail"
	// "regexp"
	// "strings"
)

// // isValidPhoneNumber_E164,
// // This returns true if a phone number is a valid e164 number, otherwise false.
// // Phone number should be sanitized before using this function. See: sanitize.Phone().
// // E164 is the international phone number standard.
// // Twilio original source: https://www.twilio.com/blog/validate-e164-phone-number-in-go
// func IsValidPhone(phone_number string) bool {
// 	e164Regex := `^\+[1-9]\d{1,14}$`
// 	re := regexp.MustCompile(e164Regex)
// 	phone_number = strings.ReplaceAll(phone_number, " ", "")
// 	return re.Find([]byte(phone_number)) != nil
// }

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
