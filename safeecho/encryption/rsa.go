package encryption

import (
	"encoding/base64"
	"fmt"
	"strings"
)

type PrivateKey struct {
	n, d *BigInt
}

type PublicKey struct {
	n, e *BigInt
}

func (p *PublicKey) encrypt(m *BigInt) *BigInt {
	return pow(m, p.e, p.n)
}

func (p *PublicKey) EncryptString(a []byte) string {
	encryptedString := make([]byte, 0)
	for i := 0; i < len(a); i++ {
		currentPart := p.encrypt(fromInt(int64(a[i])))
		encryptedString = append(encryptedString, []byte(currentPart.String())...)
		if i != len(a)-1 {
			encryptedString = append(encryptedString, []byte(",")...)
		}
	}
	return base64.StdEncoding.EncodeToString(encryptedString)
}

func (p *PublicKey) String() string {
	return fmt.Sprintf("<%s, %s>", p.n.String(), p.e.String())
}

func (p *PublicKey) Marshal() []byte {
	return []byte(fmt.Sprintf("%s,%s", p.n.String(), p.e.String()))
}

func (p *PublicKey) Unmarshal(a []byte) error {
	l := strings.Split(string(a), ",")
	p.n = fromString(l[0])
	p.e = fromString(l[1])
	return nil
}
func (p *PrivateKey) decrypt(c *BigInt) *BigInt {
	return pow(c, p.d, p.n)
}

func (p *PrivateKey) DecryptString(a string) []byte {
	encryptedArray, err := base64.StdEncoding.DecodeString(a)
	if err != nil {
		panic(err)
	}
	splitStr := strings.Split(string(encryptedArray), ",")
	decryptedString := make([]byte, 0)
	for i := 0; i < len(splitStr); i++ {
		currentPart := p.decrypt(fromString(splitStr[i]))
		decryptedString = append(decryptedString, byte(currentPart.toInt()))
	}
	return decryptedString
}

func (p *PrivateKey) String() string {
	return fmt.Sprintf("<%s, %s>", p.n.String(), p.d.String())
}

func (p *PrivateKey) Marshal() []byte {
	return []byte(fmt.Sprintf("%s,%s", p.n.String(), p.d.String()))
}

func (p *PrivateKey) Unmarshal(a []byte) error {
	l := strings.Split(string(a), ",")
	p.n = fromString(l[0])
	p.d = fromString(l[1])
	return nil
}
