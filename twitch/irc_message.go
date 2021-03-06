package twitch

import (
	"strconv"
	"strings"
)

// Message is the IRC message struct
type Message struct {
	Raw      string
	Prefix   string
	Command  string
	Params   []string
	Trailing string
	Data     struct {
		// PRIVMSG
		DisplayName  string
		UserID       int
		StreamerID   int
		StreamerName string
		Type         struct {
			Broadcaster     bool
			Moderator       bool
			Staff           bool
			Admin           bool
			GlobalModerator bool
			Subscriber      bool
		}
		Timestamp int64
		Bits      int
		Message   string
		Arguments []string

		// USERNOTICE
		NoticeType    string
		SystemMessage string
		Raid          struct {
			Login       string
			DisplayName string
			Viewers     int
		}
		GiftSub struct {
			ID          int
			Login       string
			DisplayName string
		}
		Sub struct {
			Plan     string
			PlanName string
			Months   int
		}
	}
}

func parseLine(raw string) (*Message, error) {
	raw = strings.TrimSpace(raw)
	m := &Message{Raw: raw}

	chunks := strings.Split(raw, " :")
	if len(chunks) >= 2 {
		m.Prefix = chunks[1][1:]
	}

	if raw[0] == ':' {
		chunks := strings.SplitN(raw, " ", 2)
		m.Prefix = chunks[0][1:]
		raw = chunks[1]
	}

	chunks = strings.SplitN(raw, " ", 5)
	m.Command = chunks[0]

	raw = strings.Join(chunks[1:], " ")
	if string(m.Command[0]) == "@" {
		m.Command = chunks[2]
		raw = strings.Join(chunks[2:], " ")

		m.Data.StreamerName = chunks[3][1:]

		rawData := chunks[0]
		tags := strings.Split(rawData[1:], ";")
		for _, tag := range tags {
			spl := strings.SplitN(tag, "=", 2)

			switch spl[0] {
			case "display-name":
				m.Data.DisplayName = spl[1]
				if m.Data.DisplayName == "" {
					m.Data.DisplayName = m.Nick()
				}
			case "user-id":
				m.Data.UserID, _ = strconv.Atoi(spl[1])
			case "room-id":
				m.Data.StreamerID, _ = strconv.Atoi(spl[1])
			case "mod":
				m.Data.Type.Moderator, _ = strconv.ParseBool(spl[1])
			case "subscriber":
				m.Data.Type.Subscriber, _ = strconv.ParseBool(spl[1])
			case "tmi-sent-ts":
				m.Data.Timestamp, _ = strconv.ParseInt(spl[1], 10, 64)
			case "user-type":
				switch spl[1] {
				case "staff":
					m.Data.Type.Staff = true
				case "admin":
					m.Data.Type.Admin = true
				case "global_mod":
					m.Data.Type.GlobalModerator = true
				}
			case "bits":
				m.Data.Bits, _ = strconv.Atoi(spl[1])
			case "msg-id":
				m.Data.NoticeType = spl[1]
			case "msg-param-displayName":
				m.Data.Raid.DisplayName = spl[1]
			case "msg-param-login":
				m.Data.Raid.Login = spl[1]
			case "msg-param-viewerCount":
				m.Data.Raid.Viewers, _ = strconv.Atoi(spl[1])
			case "msg-param-recipient-id":
				m.Data.GiftSub.ID, _ = strconv.Atoi(spl[1])
			case "msg-param-recipient-display-name":
				m.Data.GiftSub.DisplayName = spl[1]
			case "msg-param-recipient-user-name":
				m.Data.GiftSub.Login = spl[1]
			case "msg-param-sub-plan":
				m.Data.Sub.Plan = spl[1]
			case "msg-param-sub-plan-name":
				m.Data.Sub.PlanName = spl[1]
			case "msg-param-months":
				m.Data.Sub.Months, _ = strconv.Atoi(spl[1])
			}
		}

		if strings.ToLower(m.Data.StreamerName) == strings.ToLower(m.Data.DisplayName) {
			m.Data.Type.Broadcaster = true
		}
	}

	if raw[0] != ':' {
		chunks := strings.SplitN(raw, " :", 2)
		m.Params = strings.Split(chunks[0], " ")
		if len(chunks) == 2 {
			raw = chunks[1]
			m.Data.Message = chunks[1]
			m.Data.Arguments = strings.Split(chunks[1], " ")
		} else {
			raw = ""
		}
	}

	if len(raw) > 0 {
		if raw[0] == ':' {
			raw = raw[1:]
		}
		m.Trailing = raw
	}

	return m, nil
}

// String is the raw message from the IRC message
func (m *Message) String() string {
	return m.Raw
}

// Nick returns the Nick of the sender
func (m *Message) Nick() string {
	if m.Prefix == "" {
		return ""
	}
	return strings.SplitN(m.Prefix, "!", 2)[0]
}
