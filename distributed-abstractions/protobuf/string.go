package protobuf

import "fmt"

// toString converts a proto message to a string. Note that network message
// are printed separately.
func (m *Message) ToString() string {

	first := fmt.Sprintf("%s from %s to %s", m.Type, m.FromAbstractionId, m.ToAbstractionId)
	out := ""
	if m.Type == Message_PL_DELIVER {
		out += first + fmt.Sprintf(" containing {%v}", m.GetPlDeliver().GetMessage().ToString())
	}
	if m.Type == Message_PL_SEND {
		out += first + fmt.Sprintf(" containing {%v}", m.GetPlSend().GetMessage().ToString())
	}
	if m.Type == Message_APP_VALUE {
		out += fmt.Sprintf(" APP_VALUE(%d) from %s to %s", m.GetAppValue().GetValue().GetV(), m.GetFromAbstractionId(), m.GetToAbstractionId())
	}
	if m.Type == Message_BEB_BROADCAST {
		out += first + fmt.Sprintf(" containing {%v}", m.GetBebBroadcast().GetMessage().ToString())
	}
	return out
}
