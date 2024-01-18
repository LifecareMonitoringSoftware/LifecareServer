package model

// The document name will be the same as the email.
type User struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

const FirstNameField string = "FirstName"
const LastNameField string = "LastName"
const EmailField string = "Email"
const PhoneField string = "Phone"
