package storage

import (
	"context"
	"errors"
	"log"

	"cloud.google.com/go/firestore"
	"example.com/micro/cloud_run"
	"example.com/micro/model"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

const firestoreLocalConnectionFilePath = "" +
	"/Volumes/Local/zNotSynced/FirestoreConnectionFiles/lifecare-server-firestore-connection-fpc2l-ac5b7ca158.json"

var ctx = context.Background()
var client = getFirestoreClient()

var ErrorNotUnique = errors.New("The requested item was found more than once.")
var ErrorNotFound = errors.New("The requested item was not found.")

// getFirestoreClient, This will return a fire store client object that can be used to access the
// fire store database. When we are running on cloud run, the application default credentials will
// be used generate the fire store client. When we are not running on cloud run, this function
// requires and uses a local firestore credentials key file in a prearranged directory.
// See also: https://firebase.google.com/docs/firestore/quickstart#initialize
// This will not close the client.
// After calling this function, the caller should generally call: defer client.Close()
func getFirestoreClient() *firestore.Client {
	ctx := context.Background()
	var firebaseApp *firebase.App
	var firebaseAppError error
	if cloud_run.IsRunningOnCloudRun() {
		// We are running on cloud run.
		// Use the application default credentials to get the firestore client.
		conf := &firebase.Config{ProjectID: cloud_run.ProjectID()}
		firebaseApp, firebaseAppError = firebase.NewApp(ctx, conf)
	} else {
		// We are not running on cloud run.
		// Use a credentials key file.
		clientOption := option.WithCredentialsFile(firestoreLocalConnectionFilePath)
		firebaseApp, firebaseAppError = firebase.NewApp(ctx, nil, clientOption)
	}
	if firebaseAppError != nil {
		log.Fatalln(firebaseAppError)
	}
	client, firestoreClientError := firebaseApp.Firestore(ctx)
	if firestoreClientError != nil {
		log.Fatalln(firestoreClientError)
	}
	return client
}

func SetUser(user model.User) error {
	_, errorOrNil := client.Collection(model.UsersCL).Doc(user.Email).Set(ctx, user)
	if errorOrNil != nil {
		log.Printf("Could not set user in the database.: %s", errorOrNil)
	}
	return errorOrNil
}

func GetUserWithEmailOrNil(standardEmail string) *model.User {
	document, errorOrNil := client.Collection(model.UsersCL).Doc(standardEmail).Get(ctx)
	if errorOrNil != nil {
		return nil
	}
	var user model.User
	document.DataTo(&user)
	return &user
}

func GetSingleUserWithPhoneOrError(standardPhone string) (*model.User, error) {
	iter := client.Collection(model.UsersCL).
		Where(model.PhoneField, "==", standardPhone).Documents(ctx)
	documentArray, error := iter.GetAll()
	if error != nil {
		// Firestore error.
		return nil, error
	}
	if len(documentArray) == 0 {
		// None found.
		return nil, ErrorNotFound
	}
	if len(documentArray) > 1 {
		// More than one was found.
		// Phone numbers are supposed to be unique.
		return nil, ErrorNotUnique
	}
	var user model.User
	documentArray[0].DataTo(&user)
	return &user, nil
}
