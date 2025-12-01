package controllers

import (
	"Avocycle/config"
	// "Avocycle/models"
	// "Avocycle/utils"
	"net/http"
	"os"

	// "github.com/danilopolani/gocialite/structs"
	"github.com/gin-gonic/gin"
)

// Redirect to correct oAuth 
// Redirect to correct oAuth URL
// RedirectHandlerPetani godoc
// @Summary Login via Google OAuth (Petani)
// @Description This endpoint will redirect users to Google Sign-in page in browser.
// @Description 
// @Description ⚠ Cannot be tested directly via Swagger or Postman.
// @Description 
// @Description Please open this URL in a normal browser instead:
// @Description  
// @Description http://localhost:2005/api/v1/auth/google/petani
// @Tags Auth Petani with Google
// @Produce json
// @Success 302 {string} string "Redirect to Google OAuth"
// @Router /auth/google/petani [get]
func RedirectHandlerPetani(c *gin.Context) {
    // provider := c.Param("provider")
    provider := "google"

    providerSecrets := map[string]map[string]string{
        "google": {
            "clientID":     os.Getenv("CLIENT_ID_GOOGLE"),
            "clientSecret": os.Getenv("CLIENT_SECRET_GOOGLE"),
            "redirectURL":  os.Getenv("AUTH_REDIRECT_URL") + "/google/callback/petani",
        },
    }

    providerScopes := map[string][]string{
        "google": {},
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

    if err != nil {
        c.Writer.Write([]byte("Error: " + err.Error()))
        return
    }

    c.Redirect(http.StatusFound, authURL)
}


// Handle callback of provider
// CallbackHandlerPetani godoc
// @Summary Google OAuth Callback (Petani)
// @Description Handle Google OAuth callback and return JWT token for Petani.
// @Description 
// @Description Setelah login dengan Google, browser akan menampilkan JSON berikut:
// @Description 
// @Tags Auth Petani with Google
// @Produce json
// @Success 200 {object} map[string]interface{} "Login success" 
// @Router /auth/{provider}/callback/petani [get]
// @Description {
// @Description   "action": "google auth petani",
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
// @Description    . 	"role": "Petani"
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
// RedirectHandlerPetani tetap sama seperti sekarang
func CallbackHandlerPetani(c *gin.Context) {
    handleGoogleCallback(c, "Petani")
}

// func getOrRegisterUser(provider string, user *structs.User) models.User {
// 	var userData models.User

// 	// Get database connection
//     db, err := config.DbConnect()
//     if err != nil {
//         panic("Failed to connect database")
//     }

// 	db.Where("auth_provider = ? AND provider_id = ?", provider, user.ID).First(&userData)

// 	if userData.ID == 0{
// 		newUser := models.User{
// 			FullName: user.FullName,
// 			Email: user.Email,
// 			AuthProvider: string("Google"),
// 			ProviderID: user.ID,
// 			Role: string("Petani"),
// 		}

// 		db.Create(&newUser)
// 		return newUser
// 	} else {
// 		return userData
// 	}
// }