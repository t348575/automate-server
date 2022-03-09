package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/automate/automate-server/general-service/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/argon2"
	"golang.org/x/oauth2"
)

var (
	errInvalidHash         = errors.New("the encoded hash is not in the correct format")
	errIncompatibleVersion = errors.New("incompatible version of argon2")
)

type BaseConfig interface {
	GetPort() string
	GetTimeout() int
	GetReadBufferSize() int
	GetAppName() string
	GetIsProduction() bool
	GetCookieKey() string
	GetBodyLimit() int
}

func GenerateRandomBytes(size uint32) []byte {
	token := make([]byte, size)
	rand.Read(token)
	return token
}

func DecodeBase64(message []byte) ([]byte, error) {
	base64Text := make([]byte, base64.StdEncoding.DecodedLen(len(message)))

	_, err := base64.URLEncoding.Decode(base64Text, message)
	if err != nil {
		return nil, err
	}
	return base64Text, nil
}

func EncodeBase64(message []byte) []byte {
	base64Text := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.URLEncoding.Encode(base64Text, message)
	return base64Text
}

func ParseFlags() bool {
	devMode := flag.Bool("dev", false, "Run in dev mode")
	envFile := flag.String("env", "", ".env file path")

	flag.Parse()

	if err := godotenv.Load(func() string {
		if len(*envFile) > 0 {
			return *envFile
		}

		return ".prod.env"
	}()); err != nil {
		log.Panic().Err(err).Msg("Could not load .env file")
	}

	return !*devMode
}

func IsInList(item string, list *[]string) int {
	for i, val := range *list {
		if val == item {
			return i
		}
	}
	return -1
}

type JwtConfig struct {
	User       string
	ExpireIn   time.Duration
	Scope      string
	Subject    string
	Data       map[string]string
	PrivateKey *rsa.PrivateKey
}

func CreateJwt(c JwtConfig) (string, error) {
	now := time.Now().UTC()
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user":  c.User,
		"data":  c.Data,
		"scope": c.Scope,
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"sub":   c.Subject,
		"exp":   now.Add(c.ExpireIn).Unix(),
	}).SignedString(c.PrivateKey)

	if err != nil {
		return "", err
	}
	return token, nil
}

type params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func VerifyHash(password string, hash string) bool {
	p, salt, plainHash, err := decodeHash(hash)
	if err != nil {
		return false
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(plainHash, otherHash) == 1 {
		return true
	}
	return false
}

func decodeHash(encodedHash string) (p *params, salt, hash []byte, err error) {
	values := strings.Split(encodedHash, "$")
	if len(values) != 6 {
		return nil, nil, nil, errInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(values[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, errIncompatibleVersion
	}

	p = &params{}
	_, err = fmt.Sscanf(values[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(values[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func InterfaceToStringArray(in []interface{}) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.(string)
	}
	return out
}

func HashPassword(password string) (encodedHash string, err error) {
	salt := GenerateRandomBytes(16)

	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return "$argon2id$v=19$m=65536,t=3,p=2$" + b64Salt + "$" + b64Hash, nil
}

func OAuthJwt(user, scope string, jwtPrivateKey *rsa.PrivateKey) (*oauth2.Token, error) {
	refreshJwt, err := CreateJwt(JwtConfig{
		User:       user,
		ExpireIn:   time.Minute * 90,
		Scope:      scope,
		Subject:    "refresh",
		Data:       map[string]string{},
		PrivateKey: jwtPrivateKey,
	})
	if err != nil {
		return nil, err
	}

	accessJwt, err := CreateJwt(JwtConfig{
		User:       user,
		ExpireIn:   time.Minute * 60,
		Scope:      scope,
		Subject:    "access",
		Data:       map[string]string{},
		PrivateKey: jwtPrivateKey,
	})
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  accessJwt,
		RefreshToken: refreshJwt,
		Expiry:       time.Now().Add(time.Minute * 60),
	}, nil
}

func Format(in string, data map[string]string) string {
	for k, v := range data {
		in = strings.Replace(string(in), k, v, -1)
	}
	return in
}

func GetFromMapArray(data []map[string]string, key string, value string) int {
	for i, v := range data {
		if v[key] == value {
			return i
		}
	}
	return -1
}

func ValidateStruct(err error) []*ErrorResponse {
	var errors []*ErrorResponse
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

func SendEmail(url string, config *models.SendEmailConfig) error {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)
	req.SetRequestURI(url)
	req.Header.Set("Content-Type", "application/json")

	body, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req.SetBody(body)
	if err := a.Parse(); err != nil {
		return err
	}

	code, body, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if errArr != nil || len(errArr) != 0 {
		return errArr[0]
	}

	if code != fiber.StatusOK && code != fiber.StatusCreated {
		fmt.Println(string(body))
		return errors.New(string(body))
	}

	return nil
}

func ConvertConfig[T, S any](input T) (*S, error) {
	res, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	cfg := new(S)
	err = json.Unmarshal(res, cfg)

	return cfg, err
}
