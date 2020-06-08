package config

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

type PASSWORD struct {
	// These will be stored in sha512/hexedecimal
	Digest string `json:"digest"`
	Salt   string `json:"salt"`
}

func (p *PASSWORD) Set(password string) error {
	newsalt, err := GenerateRandomBytes(32)
	checkErr(err)

	hasher := sha1.New()
	hasher.Write(newsalt)
	hasher.Write([]byte(password))

	p.Digest = hex.EncodeToString(hasher.Sum(nil))
	p.Salt = hex.EncodeToString(newsalt)

	return nil
}

func (p *PASSWORD) Test(password string) bool {
	salt, err := hex.DecodeString(p.Salt)
	checkErr(err)

	hasher := sha1.New()
	hasher.Write(salt)
	hasher.Write([]byte(password))

	return p.Digest == hex.EncodeToString(hasher.Sum(nil))
}

func (p *PASSWORD) Serialize() string {
	s, e := json.Marshal(p)
	checkErr(e)
	return string(s)
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
