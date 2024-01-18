package main

import (
	"fmt"
	"html/template"
	"net/http"

	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"example.com/micro/model"
	"example.com/micro/sanitize"
	"example.com/micro/storage"
	"example.com/micro/validate"
	// "golang.org/x/crypto/bcrypt"
)

// Directories and files.
const staticDirectory string = "web_static"
const signupFilePath string = "web_templates/signup.html"

// Dynamic Routes.
const rootRoute string = "/"
const staticRoute string = "/page"
const signupRoute string = "/signup"
const adminRoute string = "/admin"

// Static routes, for reference only.
// termsOfServiceRoute = "/terms"
// privacyPolicyRoute = "/privacy"

type SignupPageFields struct {
	MessageToUser      string
	FirstName          string
	LastName           string
	Phone              string
	Email              string
	MessagingConsent   string
	TermsAndConditions string
	Disclaimer         string
}

func createRouterAndRoutes() http.Handler {
	routerMain := chi.NewRouter()

	// Protected routes
	routerMain.Group(func(router chi.Router) {
		// Seek, verify and validate JWT tokens
		router.Use(jwtauth.Verifier(tokenConfig))

		// Handle valid / invalid tokens. In this example, we use
		// the provided authenticator middleware, but you can write your
		// own very easily, look at the Authenticator method in jwtauth.go
		// and tweak it, its not scary.
		router.Use(jwtauth.Authenticator(tokenConfig))

		router.Get(adminRoute, func(w http.ResponseWriter, r *http.Request) {
			_, claims, _ := jwtauth.FromContext(r.Context())
			w.Write([]byte(fmt.Sprintf("Protected area. Hi %v", claims["user_id"])))
		})
	})

	// Public routes
	routerMain.Group(func(router chi.Router) {
		router.Get(rootRoute, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello World"))
		})

		router.Get(signupRoute, doSignupPage)
		router.Post(signupRoute, doSignupPage)
	})

	// Create a route for /static that will serve contents from the static html folder.
	workingDirectory, _ := os.Getwd()
	fullStaticDirectory := http.Dir(filepath.Join(workingDirectory, staticDirectory))
	FileServer(routerMain, staticRoute, fullStaticDirectory)

	return routerMain
}

// The FileServer function sets up a http.FileServer handler to serve static files
// from an http.FileSystem.
func FileServer(router chi.Router, urlPath string, rootDirectory http.FileSystem) {
	if strings.ContainsAny(urlPath, "{}*") {
		panic("The static FileServer does not permit any URL parameters.")
	}

	if urlPath != "/" && urlPath[len(urlPath)-1] != '/' {
		router.Get(urlPath, http.RedirectHandler(
			urlPath+"/", http.StatusMovedPermanently).ServeHTTP)
		urlPath += "/"
	}
	urlPath += "*"

	router.Get(urlPath, func(w http.ResponseWriter, r *http.Request) {
		routeContext := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(routeContext.RoutePattern(), "/*")
		handler := http.StripPrefix(pathPrefix, http.FileServer(rootDirectory))
		handler.ServeHTTP(w, r)
	})
}

func doSignupPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		data := SignupPageFields{
			MessageToUser:      "",
			FirstName:          "",
			LastName:           "",
			Phone:              "",
			Email:              "",
			MessagingConsent:   "",
			TermsAndConditions: "",
			Disclaimer:         "",
		}
		signupTemplate := template.Must(template.ParseFiles(signupFilePath))
		signupTemplate.Execute(w, data)
		return
	}

	// Get the form fields.
	firstName := r.FormValue("FirstName")
	lastName := r.FormValue("LastName")
	email := r.FormValue("Email")
	phone := r.FormValue("Phone")
	messagingConsent := r.FormValue("MessagingConsent")
	termsAndConditions := r.FormValue("TermsAndConditions")
	disclaimer := r.FormValue("Disclaimer")

	// Sanitize the form fields.
	firstName = sanitize.NameProperSingle(firstName)
	lastName = sanitize.NameProperSingle(lastName)
	phone = sanitize.Phone(phone)
	email = sanitize.Email(email, false)

	// Check for blank fields.
	if firstName == "" {
		showSignupPageWithErrorMessage(w, r, "Please enter a first name.")
		return
	}
	if lastName == "" {
		showSignupPageWithErrorMessage(w, r, "Please enter a last name.")
		return
	}
	if phone == "" {
		showSignupPageWithErrorMessage(w, r, "Please enter a phone number.")
		return
	}
	if email == "" {
		showSignupPageWithErrorMessage(w, r, "Please enter an email address.")
		return
	}
	// Check for a valid email and phone number.
	if !validate.IsValidEmail(email) {
		showSignupPageWithErrorMessage(w, r, "Please enter a valid email address.")
		return
	}
	// We're skipping the phone number validation for now, because phone validation can be difficult.
	// We might validate simply with verification later, by sending a text.
	// Validation might be possible with a Google phone number app, but you might need to figure out
	// or ask for the users country code.
	// if !validate.IsValidPhone_E164(phone) {
	// 	showSignInPageWithMessage(writer, "Please enter a valid phone number.")
	// 	return
	// }
	//
	// Ensure that all the checkboxes are selected.
	if messagingConsent != "on" ||
		termsAndConditions != "on" ||
		disclaimer != "on" {
		showSignupPageWithErrorMessage(w, r, "To create an account, please check all of the checkbox items to indicate your agreement.")
		return
	}

	// Check for duplicate users.
	userWithEmailOrNil := storage.GetUserWithEmailOrNil(email)
	// Note: Since the email field is the document ID, we don't have to check for a NotUniqueError.
	if userWithEmailOrNil != nil {
		showSignupPageWithErrorMessage(w, r, "A user with that email address already exists.")
		return
	}
	userWithPhone, error := storage.GetSingleUserWithPhoneOrError(phone)
	if userWithPhone != nil || error == storage.ErrorNotUnique {
		showSignupPageWithErrorMessage(w, r, "A user with that phone number already exists.")
		return
	}
	user := model.User{FirstName: firstName, LastName: lastName, Phone: phone, Email: email}
	errorOrNil := storage.SetUser(user)
	if errorOrNil != nil {
		showSignupPageWithErrorMessage(w, r, "Could not create user. The error was:\n"+
			errorOrNil.Error())
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Account creation was successful."))
}

func showSignupPageWithErrorMessage(w http.ResponseWriter, r *http.Request, message string) {
	// Get the form fields.
	firstName := r.FormValue("FirstName")
	lastName := r.FormValue("LastName")
	phone := r.FormValue("Phone")
	email := r.FormValue("Email")
	data := SignupPageFields{
		MessageToUser: message,
		FirstName:     firstName,
		LastName:      lastName,
		Phone:         phone,
		Email:         email,
	}
	w.WriteHeader(http.StatusBadRequest)
	signupTemplate := template.Must(template.ParseFiles(signupFilePath))
	signupTemplate.Execute(w, data)
}
