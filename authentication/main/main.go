package main
// the go language define variables inside 2(), just to remeber it
import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"time"
	"weatherapp.com/auth/authdb" //here import authdb.go file , inclusde controller

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const ( // thsi default for env variables , but i will comment them to check i can get them from the container
	defaultDBHost     = "127.0.0.1" // will be dns of mysql headless service the it will resolve to ip of pod of statefulset that run mysql
	defaultDBUser     = "authuser" // will be gaber
	defaultDBPassword = "authpassword" // will be gaber123
	defaultDBName     = "weatherapp" // will be weatherapp_db
	defaultDBPort     = "3306" //will be same
	defaultSecretKey  = "xco0sr0fh4e52x03g9mv" //will be secret 123 
	defaultAuthPort   = "8080" // this is the port this servcie run on from inside code 
)

type Token struct {
	Role        string `json:"role"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}
// this servcie need these env variables so need to pass it to the container when we run the container
var (
	dbHost     = getEnv("DB_HOST", defaultDBHost)
	dbUser     = getEnv("DB_USER", defaultDBUser)
	dbPassword = getEnv("DB_PASSWORD", defaultDBPassword)
	dbName     = getEnv("DB_NAME", defaultDBName)
	dbPort     = getEnv("DB_PORT", defaultDBPort)
	secretKey  = getEnv("SECRET_KEY", defaultSecretKey) //secret123 will extract from secret and inject to the container as env variable and use it to generate jwt when login
	authPort   = getEnv("AUTH_PORT", defaultAuthPort)
)

func main() { // this is the main function
	//here use authdb.go file to connect to db and do operations on db
	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		fmt.Println(err.Error())
	}
	// authdb.CreateDB(db, dbName)
	authdb.CreateTables(db, dbName)
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowMethods("OPTIONS")
	router.Use(cors.New(corsConfig))
	router.GET("/", health)
	router.POST("/users/:id", loginUser)
	router.POST("/users", createUser)
	router.Run(":" + authPort)
}
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

type UserCreds struct {
	Username string `json:"user_name"`
	Password string `json:"user_password"`
}

func health(c *gin.Context) {
	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		fmt.Println(err.Error())
	}
	if db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not connect to the database"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": "The auth is running"})
	}
}

func loginUser(c *gin.Context) { // login 
	var uc UserCreds
	err := c.BindJSON(&uc)
	if err != nil {
		fmt.Println("Received invalid JSON for user login")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Incorrect or invalid JSON"})
		return
	}
	encPasswordb := md5.Sum([]byte(uc.Password))
	encPassword := hex.EncodeToString(encPasswordb[:])
	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		fmt.Println(err.Error())
	}
	u, err := authdb.GetUserByName(uc.Username, db, dbName)
	if err != nil {
		fmt.Println(err)
	}
	if u != (authdb.User{}) && u.Password == encPassword {
		token, err := GenerateJWT(u.Name)
		if err != nil {
			fmt.Println("Error while generating the token:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"JWT": token})
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bad credentials"})
	}
}
func createUser(c *gin.Context) {
	var u authdb.User
	c.BindJSON(&u)
	db, err := authdb.Connect(dbUser, dbPassword, dbHost, dbPort)
	if err != nil {
		fmt.Println(err.Error())
	}
	result, err := authdb.CreateUser(db, u, dbName)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while adding the user. Please check the logs"})
		return
	} else if !result {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "User already exists"})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"success": "User added successfully"})
	}
}
func GenerateJWT(userName string) (string, error) {
	var mySigningKey = []byte(secretKey) // this secret key is env variable and inject this env variable to the container and extract it from secret on k8s cluster
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = userName
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil
}
