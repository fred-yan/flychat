package api

import (
	"encoding/json"
	"errors"
	"flychat/lib"
	"flychat/model"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Auth struct{}

func (a *Auth) Signup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 输入验证
	if req.Username == "" || !isValidEmail(req.Email) || !isValidPassword(req.Password) {
		http.Error(w, "Validation error", http.StatusBadRequest)
		return
	}

	// 唯一性检查
	if userExists(req.Username, req.Email) {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 存储用户信息
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	if err := createUser(user); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isValidPassword(password string) bool {
	// 密码最小长度
	const minLen = 8
	// 密码最大长度
	const maxLen = 64
	// 是否包含数字
	hasNumber := false
	// 是否包含小写字母
	hasLower := false
	// 是否包含大写字母
	hasUpper := false
	// 是否包含特殊字符
	hasSpecial := false

	// 空字符串检查
	if len(password) == 0 {
		return false
	}

	// 长度检查
	if len(password) < minLen || len(password) > maxLen {
		return false
	}

	// 字符类型检查
	for _, char := range password {
		switch {
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 至少包含数字、小写字母、大写字母、特殊字符中的三种
	if boolToInt(hasNumber)+boolToInt(hasLower)+boolToInt(hasUpper)+boolToInt(hasSpecial) < 3 {
		return false
	}

	return true
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// isValidEmail 检查给定的电子邮件地址是否有效
// 参数:
//
//	email: 需要验证的电子邮件地址
//
// 返回:
//
//	如果电子邮件地址有效，返回 true；否则返回 false
func isValidEmail(email string) bool {
	if strings.TrimSpace(email) == "" {
		return false
	}
	return emailRegex.MatchString(email)
}

func (a *Auth) Signin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 输入验证
	if req.Username == "" || req.Password == "" {
		http.Error(w, "Validation error", http.StatusBadRequest)
		return
	}

	// 用户验证
	user, err := getUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// 生成令牌
	token, err := generateJWT(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing authorization header", http.StatusUnauthorized)
		return
	}

	tokenStr := strings.Replace(authHeader, "Bearer ", "", 1)
	if tokenStr == "" {
		http.Error(w, "Invalid token format", http.StatusUnauthorized)
		return
	}

	// 验证令牌
	//claims, err := validateJWT(tokenStr)
	//if err != nil {
	//	http.Error(w, "Invalid token", http.StatusUnauthorized)
	//	return
	//}

	// 标记令牌为无效
	if err := blacklistToken(tokenStr); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func getUserByUsername(username string) (*model.User, error) {
	var user model.User
	db := lib.DB
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &user, nil
}

func createUser(user *model.User) error {
	db := lib.DB
	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func userExists(username, email string) bool {
	var count int64
	db := lib.DB
	if err := db.Model(&model.User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error; err != nil {
		log.Printf("Failed to check user existence: %v", err)
		return false
	}
	return count > 0
}

func generateJWT(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenStr, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}

func validateJWT(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("your-secret-key"), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func blacklistToken(tokenStr string) error {
	blacklistedToken := BlacklistedToken{Token: tokenStr}
	db := lib.DB
	if err := db.Create(&blacklistedToken).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}
	return nil
}

type BlacklistedToken struct {
	Token string `gorm:"unique"`
}
