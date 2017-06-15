package handlers

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn/models"
	"github.com/keratin/authn/services"
)

type request struct {
	Username string
	Password string
}

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	// Create the account
	account, errors := services.AccountCreator(
		app.AccountStore,
		&app.Config,
		req.FormValue("username"),
		req.FormValue("password"),
	)
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	// Create the session token
	session, err := models.NewSessionJWT(
		app.RefreshTokenStore,
		app.Config,
		account.Id,
	)
	if err != nil {
		panic(err)
	}

	// Create the identity token
	identity, err := models.NewIdentityJWT(
		app.RefreshTokenStore,
		app.Config,
		session,
	)
	if err != nil {
		panic(err)
	}

	// Begin the response
	w.WriteHeader(http.StatusCreated)

	// Return the signed session in a cookie
	sessionString, err := session.Sign(jwt.SigningMethodHS256, app.Config.SessionSigningKey)
	if err != nil {
		panic(err)
	}
	sessionCookie := http.Cookie{
		Name:     "authn",
		Value:    sessionString,
		Path:     app.Config.MountedPath,
		Secure:   app.Config.ForceSSL,
		HttpOnly: true,
	}
	http.SetCookie(w, &sessionCookie)

	// Return the identity token in the body
	identityString, err := identity.Sign(jwt.SigningMethodRS256, app.Config.IdentitySigningKey)
	if err != nil {
		panic(err)
	}
	writeData(w, response{identityString})
}
