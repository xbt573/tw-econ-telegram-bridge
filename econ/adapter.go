package econ

import "regexp"

var (
	teeworldsChatRegex  = regexp.MustCompile(`\[chat\]: \d+:-?\d+:(.*)`)
	teeworldsJoinRegex  = regexp.MustCompile(`\[game\]: team_join player='\d+:(.*)'.*=\d+`)
	teeworldsLeaveRegex = regexp.MustCompile(`\[game\]: leave player='\d+:(.*)'`)

	trainfngChatRegex = regexp.MustCompile(`\[.*?\]\[chat\]: \d+:-?\d+:(.*)`)
	trainfngJoinRegex = regexp.MustCompile(`\[.*\]\[.*\]: \*\*\* '(.*)' (.*)`)

	ddnetChatRegex = regexp.MustCompile(`.* I chat: \d+:-?\d+:(.*)`)
	ddnetJoinRegex = regexp.MustCompile(`.* I chat: \*\*\* '(.*)' (.*)`)
)

type ServerType string

const (
	TEEWORLDS ServerType = "teeworlds"
	TRAINFNG             = "trainfng"
	DDNET                = "ddnet"
)

type Adapter interface {
	Match([]byte) (string, bool)
}

var Adapters = map[ServerType]Adapter{
	TEEWORLDS: teeworldsAdapter{},
	TRAINFNG:  trainfngAdapter{},
	DDNET:     ddnetAdapter{},
}

type teeworldsAdapter struct{}

func (t teeworldsAdapter) Match(bytes []byte) (string, bool) {
	match := teeworldsChatRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1], true
	}

	match = teeworldsJoinRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1] + " joined the game", true
	}

	match = teeworldsLeaveRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1] + " left the game", true
	}

	return "", false
}

type trainfngAdapter struct{}

func (trainfngAdapter) Match(bytes []byte) (string, bool) {
	match := trainfngChatRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1], true
	}

	match = trainfngJoinRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1] + " " + match[2], true
	}

	return "", false
}

type ddnetAdapter struct{}

func (ddnetAdapter) Match(bytes []byte) (string, bool) {
	match := ddnetChatRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1], true
	}

	match = ddnetJoinRegex.FindStringSubmatch(string(bytes))
	if len(match) != 0 {
		return match[1] + " " + match[2], true
	}

	return "", false
}
