package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"time"

	"github.com/gin-gonic/gin"
	"systementor.se/godemosite/data"

	//1
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	_ "github.com/gin-contrib/sessions/redis"
	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors"
)

type PageView struct {
	CurrentUser string
	PageTitle   string
	Title       string
	Text        string
}

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

func start(c *gin.Context) {
	fmt.Println("In Start")
	session := sessions.Default(c)
	user := session.Get(userkey)
	var currentUser = ""
	if user != nil {
		currentUser = user.(string)
	}
	computerName, _ := os.Hostname()
	c.HTML(http.StatusOK, "home.html", &PageView{CurrentUser: currentUser, PageTitle: "test", Title: "Hej Golang", Text: computerName})
}

func secretfunc(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	var currentUser = ""
	if user != nil {
		currentUser = user.(string)
	}
	c.HTML(http.StatusOK, "secret.html", &PageView{CurrentUser: currentUser, PageTitle: "test", Title: "Hej Golang", Text: "hejsan"})
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

	// username := c.PostForm("Email")
	// password := c.PostForm("password")
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
	session := sessions.Default(c)

	c.ShouldBindJSON(&cred)

	fmt.Println("password:")
	fmt.Println(cred.Password)
	fmt.Println("email:")
	fmt.Println(cred.Email)
	fmt.Println(cred)

	var user data.User
	data.DB.Where("Email = ?", cred.Email).First(&user)

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cred.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		fmt.Println("Wrong password")
		return
	}

	token, _ := CreateToken(user.Id)

	session.Set("user_id", user.Id)
	session.Save()
	fmt.Println("Returning json at end of function")
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
	//2
	var secret = []byte("secret")
	// store, _ := redis.NewStore(10, "tcp", config.Redis.Server, "", secret)
	store := memstore.NewStore([]byte(secret))
	router.Use(sessions.Sessions("mysession2", store))

	router.LoadHTMLGlob("templates/**")
	router.GET("/", start)
	router.GET("/login", login)
	router.POST("/login", loginPost)
	router.GET("/logout", logout)
	router.GET("/new_user", new_user)
	router.POST("/new_user", newUserPost)

	//3 frivillig
	adminRoutes := router.Group("/admin")
	adminRoutes.Use(AuthRequired)
	adminRoutes.GET("/account", secretfunc)

	router.Run(":8080")
}

// 4 frivillig
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(userkey)
	fmt.Println("user:")
	fmt.Println(user)
	var redirectUrl = url.QueryEscape("http://" + c.Request.Host + c.Request.RequestURI)
	if user == nil {
		c.Redirect(302, "/login?redirect_uri="+redirectUrl)
		// Abort the request with the appropriate error code
		return
	}
	fmt.Println("Authenticated")
	// Continue down the chain to handler etc
	c.Next()
}
