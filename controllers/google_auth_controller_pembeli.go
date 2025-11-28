package controllers

import (
	"Avocycle/config"
	"Avocycle/models"
	"Avocycle/utils"
	"net/http"
	"os"

	"github.com/danilopolani/gocialite/structs"
	"github.com/gin-gonic/gin"
)

// Redirect to correct oAuth URL
// RedirectHandlerPembeli godoc
// @Summary Login via Google OAuth (Pembeli)
// @Description This endpoint will redirect users to Google Sign-in page in browser.
// @Description 
// @Description ⚠ Cannot be tested directly via Swagger or Postman.
// @Description 
// @Description Please open this URL in a normal browser instead:
// @Description 
// @Description http://localhost:2005/api/v1/auth/google/pembeli
// @Tags Auth Pembeli with Google
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Router /auth/google/pembeli [get]
func RedirectHandlerPembeli(c *gin.Context) {
	// Retrieve provider from route
	provider := c.Param("provider")

	// In this case we use a map to store our secrets, but you can use dotenv or your framework configuration
	// for example, in revel you could use revel.Config.StringDefault(provider + "_clientID", "") etc.
	providerSecrets := map[string]map[string]string{
		"google": {
			"clientID":     os.Getenv("CLIENT_ID_GOOGLE"),
			"clientSecret": os.Getenv("CLIENT_SECRET_GOOGLE"),
			"redirectURL":  os.Getenv("AUTH_REDIRECT_URL") + "/google/callback/pembeli",
		},
	}

	providerScopes := map[string][]string{
		"google": []string{},
	}

	providerData := providerSecrets[provider]
	actualScopes := providerScopes[provider]
	authURL, err := config.Gocial.New().
		Driver(provider).
		Scopes(actualScopes).
		Redirect(
			providerData["clientID"],
			providerData["clientSecret"],
			providerData["redirectURL"],
		)

	// Check for errors (usually driver not valid)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	// Redirect with authURL
	c.Redirect(http.StatusFound, authURL)
}

// CallbackHandlerPembeli godoc
// @Summary Google OAuth Callback (Pembeli)
// @Description Handle Google OAuth callback and return JWT token for Pembeli.
// @Description 
// @Description Setelah login dengan Google, browser akan menampilkan JSON berikut:
// @Description 
// @Tags Auth Pembeli with Google
// @Produce json
// @Success 200 {object} map[string]interface{} "Login success" 
// @Router /auth/{provider}/callback/pembeli [get]
// @Description {
// @Description   "action": "google auth pembeli",
// @Description   "data": {
// @Description    . 	"ID": 0,
// @Description    . 	"CreatedAt": "2025-11-27T23:00:09.5797085-08:00",
// @Description    . 	"UpdatedAt": "2025-11-27T23:00:09.5797085-08:00",
// @Description    . 	"DeletedAt": null,
// @Description    . 	"fullname": "John Doe",
// @Description    .     "phone": "",
// @Description    . 	"email": "test123@gmail.com",
// @Description    . 	"password": "",
// @Description    . 	"auth_provider": "Google",
// @Description    . 	"provider_id": "110xxxxxxxxxxx",
// @Description    . 	"role": "Pembeli"
// @Description   .		},
// @Description   "jwtToken": "eyJhbGciOiJIUzI1NiI....",
// @Description   "success": true,
// @Description   "token_google": {
// @Description    . 	"access_token": "ya29.A0ATi6K....",
// @Description    . 	"token_type": "Bearer",
// @Description    . 	"expiry": "2025-11-28T00:00:08.0994068-08:00",
// @Description    . 	"expires_in": 3599
// @Description    .    }
// @Description     }
func CallbackHandlerPembeli(c *gin.Context) {
	// Retrieve query params for state and code
	state := c.Query("state")
	code := c.Query("code")
	provider := c.Param("provider")

	// Handle callback and check for errors
	user, token, err := config.Gocial.Handle(state, code)
	if err != nil {
		c.Writer.Write([]byte("Error: " + err.Error()))
		return
	}

	// create or register new user 
	newUser := getOrRegisterUserPembeli(provider, user)

	// create jwt token
	jwtToken, err := utils.GenerateJWT(&newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error" : "Failed to Generate token"})
		return
	}

	// // Print in terminal user information
	// fmt.Printf("%#v", token)
	// fmt.Printf("%#v", user)
	// fmt.Printf("%#v", provider)

	// // If no errors, show provider name
	// c.Writer.Write([]byte("Hi, " + user.FullName))
	c.JSON(http.StatusOK, gin.H{
		"success" : true,
		"action" : "google auth pembeli",
		"data" : newUser,
		"token_google" : token,
		"jwtToken" : jwtToken,
	})
}

func getOrRegisterUserPembeli(provider string, user *structs.User) models.User {
	var userData models.User

	// Get database connection
    db, err := config.DbConnect()
    if err != nil {
        panic("Failed to connect database")
    }

	db.Where("auth_provider = ? AND provider_id = ?", provider, user.ID).First(&userData)

	if userData.ID == 0{
		newUser := models.User{
			FullName: user.FullName,
			Email: user.Email,
			AuthProvider: string("Google"),
			ProviderID: user.ID,
			Role: string("Pembeli"),
		}

		db.Create(&newUser)
		return newUser
	} else {
		return userData
	}
}