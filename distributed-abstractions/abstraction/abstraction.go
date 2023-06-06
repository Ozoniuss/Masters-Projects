package abstraction

import (
	pb "hw/protobuf"
	"strings"
)

const APP string = "app"
const APP_PL string = "app.pl"
const APP_BEB string = "app.beb"
const APP_BEB_PL string = "app.beb.pl"
const APP_NNAR string = "app.nnar"

type Abstraction interface {
	Handle(m *pb.Message) error
}

// Previous returns the parent abstraction id
func Previous(abstractionId string) string {
	all := strings.Split(abstractionId, ".")
	if len(all) == 1 {
		panic("no previous abstraction")
	}
	// Trim last abstraction
	all = all[:len(all)-1]
	return strings.Join(all, ".")
}

// Next returns the Next abstraction id
func Next(abstractionId string, next string) string {
	b := strings.Builder{}
	b.WriteString(abstractionId)
	b.WriteByte('.')
	b.WriteString(next)
	return b.String()
}
