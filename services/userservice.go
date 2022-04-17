package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/beans"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/config"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/models"
	"github.com/Poojan-patel/Banking-App-Golang-ElasticSearch/repository"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/dgrijalva/jwt-go"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var secret_salt = []byte("the_banking_system")

func createUserAccount(acc *beans.Account) (*beans.User, error) {
	pass := strings.Replace(uuid.New().String(), "-", "", -1)
	user := &beans.User{
		UserId: acc.UserId,
		Mobile: acc.Mobile,
		Password: generateHash(pass),
	}
	err := saveUser(user, "")
	if err != nil {
		return nil, err
	}
	user.Password = pass
	return user, nil
}

func generateHash(plain string) string{
	hashed, _ := bcrypt.GenerateFromPassword([]byte(plain), 10)
	return string(hashed)
}

func validatePassword(hashed string, plain string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	return err == nil
}

func saveUser(usr *beans.User, id string) (error) {
	var es *elasticsearch.Client = config.GetESClient()
	return repository.Upsert(usr, "users", id, es)
	
}

func FindUser(userId *string, password *string) (map[string]interface{}, error){
	var es *elasticsearch.Client = config.GetESClient()
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":{"match_phrase":{"user_id":%s}}}}}`, *userId)
	res, err := repository.SearchQuery(&query_string, "users", es)
	if(err != nil){
		return nil, err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&resMapper)
	for _, hit := range resMapper["hits"].(map[string]interface{})["hits"].([]interface{}){
		doc := hit.(map[string]interface{})
		usr := doc["_source"].(map[string]interface{})
		if validatePassword(usr["password"].(string), *password) {
			return doc, nil
		}
	}
	return nil, nil
}

func ChangePassword(req *models.UserChangePassword) error {
	doc, err := FindUser(&req.UserId, &req.OldPassword)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("User Doesn't Exists")
	}
	id := doc["_id"].(string)
	usr := doc["_source"].(map[string]interface{})
	updated_user := &beans.User{UserId: usr["user_id"].(string), Mobile: usr["mobile"].(string), Password:generateHash(req.NewPassword)}
	return saveUser(updated_user, id)
}

func GenerateAuthToken(usrLogin *models.UserLogin) (*string, error){
	doc, err := FindUser(&usrLogin.UserId, &usrLogin.Password)
	if err != nil {
		log.Println("gat1:", err)
		return nil, err
	}
	if doc == nil {
		log.Println("gat2:", err)
		return nil, errors.New("User doesn't Exists")
	}
	claim := &models.JWTClaim{
		UserId: usrLogin.UserId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
	}
	authToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	token, err := authToken.SignedString(secret_salt)
	if err != nil {
		log.Println("gat3:", err)
		return nil, err
	} else{
		err = saveJWTToken(&usrLogin.UserId, &token)
		if err != nil {
			log.Println("gat4:", err)
			return nil, err
		}
		return &token, nil
	}

}

func ValidateAuthToken(userId *string, authToken *string) error {
	tkn, err := jwt.ParseWithClaims(*authToken, &models.JWTClaim{}, func(t *jwt.Token) (interface{}, error) {
		return secret_salt, nil
	})
	if err != nil || !tkn.Valid {
		return errors.New("Looks Like your Token is Expired or Not Valid")
	}
	es := config.GetESClient()
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":[{"match_phrase":{"user_id":"%s"}}, {"match_phrase":{"token":"%s"}}]}}}`, *userId, *authToken)
	res, err := repository.SearchQuery(&query_string, "jwt", es)
	if err != nil {
		return err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&resMapper)
	hits := resMapper["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return errors.New("Unauthorized Access, Remove the Authentication Header from request and Generate a new AuthToken")
	}
	return nil
}

func saveJWTToken(userId *string, token *string) error{
	es := config.GetESClient()
	jwttoken := &models.JWTToken{
		UserId: *userId,
		Token: *token,
	}
	query_string := fmt.Sprintf(`{"query":{"bool":{"must":{"match_phrase":{"user_id":"%s"}}}}}`, *userId)
	res, err := repository.SearchQuery(&query_string, "jwt", es)
	if err != nil {
		log.Println("sjt:", err)
		return err
	}
	var resMapper map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&resMapper)
	id := ""
	for _, hit := range resMapper["hits"].(map[string]interface{})["hits"].([]interface{}){
		doc := hit.(map[string]interface{})
		id = doc["_id"].(string)
		break
	}
	err = repository.Upsert(jwttoken, "jwt", id, es)
	return err
}