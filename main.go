// ### Server guiding rules:
// - Keep server stateless. This is important for future redundancy and availability setups.
//
// ### Code examples for technologies used by this class.
//
// Go language project template for Google Cloud Run:
// https://github.com/GoogleCloudPlatform/cloud-run-microservice-template-go
//
// JWT authorization example:
// https://raw.githubusercontent.com/go-chi/jwtauth/master/_example/main.go
//
// BCrypt Hashing example
// https://medium.com/@jcox250/password-hash-salt-using-golang-b041dc94cb72
//
// Registration and Login example
// https://github.com/xDinomode/Go-Signup-Login-Example-MySQL/blob/master/signup.go
//
// Firestore quickstart example
// https://github.com/GoogleCloudPlatform/golang-samples/blob/main/firestore/firestore_quickstart/main.go
// https://firebase.google.com/docs/firestore/data-model
// https://firebase.google.com/docs/firestore/manage-data/add-data
//
// Twilio SMS
// https://www.twilio.com/docs/messaging/tutorials/send-messages-with-messaging-services

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"example.com/micro/cloud_run"
	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// tokenConfig, This is our JWT token configuration.
var tokenConfig *jwtauth.JWTAuth

func main() {
	// Set the server port.
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Print some information about our environment to the log file.
	log.Printf("projectId: '%v'", cloud_run.ProjectID())
	log.Printf("isRunningOnCloudRun: '%v'", cloud_run.IsRunningOnCloudRun())
	log.Printf("port: %v", port)

	// Create our JSON Web Token (JWT) configuration.
	// This should be done before initializing our server or router, so that the pointer is never nil.
	tokenConfig = jwtauth.New("HS256", []byte("secret"), nil, jwt.WithAcceptableSkew(30*time.Second))

	// Create our server struct.
	var server = http.Server{
		// Add some defaults, should be changed to suit your use case.
		Addr:           ":" + port,
		ReadTimeout:    2 * 60 * time.Second,
		WriteTimeout:   2 * 60 * time.Second,
		MaxHeaderBytes: 1 << 20, // This equals 1,048,576, aka 1 megabyte.
	}

	// Setup our request router.
	server.Handler = createRouterAndRoutes()

	// Start the server, asynchronously.
	log.Println("starting HTTP server")
	go func() {
		if listenError := server.ListenAndServe(); listenError != nil && listenError != http.ErrServerClosed {
			log.Fatalf("server closed unexpectedly: %v", listenError)
		}
	}()

	// Run any rest code here.
	runTestCode()

	// Listen for SIGINT to gracefully shutdown.
	signalContext, stopFunction := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stopFunction()
	<-signalContext.Done()
	log.Println("shutdown initiated")

	// Cloud Run gives apps 10 seconds to shutdown. See
	// https://cloud.google.com/blog/topics/developers-practitioners/graceful-shutdowns-cloud-run-deep-dive
	// for more details.
	ctx, cancelFunction := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunction()
	server.Shutdown(ctx)
	log.Println("shutdown completed")
}

func runTestCode() {

	// TEST ONLY: Print a sample web token.
	// For debugging/example purposes, we generate and print
	// a sample jwt token with claims `user_id:123` here:
	_, tokenString, _ := tokenConfig.Encode(map[string]interface{}{"user_id": 123})
	fmt.Printf("DEBUG: a sample jwt is %s\n", tokenString)

}
