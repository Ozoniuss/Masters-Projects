package abstraction

import pb "hw/protobuf"

const APP string = "app"
const APP_PL string = "app.pl"
const APP_BEB string = "app.beb"
const APP_BEB_PL string = "app.beb.pl"

type Abstraction interface {
	Handle(m *pb.Message)
}
