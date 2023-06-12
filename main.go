package main

import (
	"fmt"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"systementor.se/godemosite/data"

	//1
	"github.com/gin-contrib/sessions"
	_ "github.com/gin-contrib/sessions/redis"
	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
)

type LoginView struct {
	CurrentUser string
	PageTitle   string
	Error       bool
	Email       string
}

type Credential struct {
	Email    string
	Password string
}

var userkey = "SESSION_KEY_USERID"
var jwtKey = []byte("SECRET_KEY")

func CreateToken(userId int) (string, error) {
	var err error
	// Creating Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["sub"] = userId
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix() // Token expires after 15 minutes
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return token, nil
}

func logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete(userkey)
	session.Save()
	c.Redirect(302, "/")

}

func new_user(c *gin.Context) {
	c.HTML(http.StatusOK, "new_user.html", &LoginView{PageTitle: "Create User"})
}

func newUserPost(c *gin.Context) {
	var cred Credential
	c.ShouldBindJSON(&cred)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(cred.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to hash password"})
		return
	}
	user := data.User{
		Email:    cred.Email,
		Password: string(passwordHash),
	}
	data.DB.Create(&user)
	c.JSON(http.StatusOK, "Logged in")
}

func login(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", &LoginView{PageTitle: "Login"})
}
func loginPost(c *gin.Context) {
	fmt.Println("LoginPost")
	var cred Credential

	c.ShouldBindJSON(&cred)
	var user data.User
	data.DB.Where("Email = ?", cred.Email).First(&user)

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cred.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		fmt.Println("Wrong password")
		return
	}

	token, _ := CreateToken(user.Id)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

var config Config

func main() {
	readConfig(&config)

	data.InitDatabase(config.Database.File,
		config.Database.Server,
		config.Database.Database,
		config.Database.Username,
		config.Database.Password,
		config.Database.Port)

	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	corsConfig.AllowMethods = []string{"GET", "POST"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type"}

	router.Use(cors.New(corsConfig))
	router.GET("/login", login)
	router.POST("/login", loginPost)
	router.GET("/logout", logout)
	router.GET("/new_user", new_user)
	router.POST("/new_user", newUserPost)
	router.Run(":8080")
}
