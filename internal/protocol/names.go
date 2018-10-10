package protocol

import "regexp"

var validTopicChannelNameRegex = regexp.MustCompile(`^[\.a-zA-Z0-9_-]+(#ephemeral)?$`)

// TopicName是否合法
func IsValidTopicName(name string) bool {
	return isValidName(name)
}

// ChannelName是否合法
func IsValidChannelName(name string) bool {
	return isValidName(name)
}

func isValidName(name string) bool {
	if len(name) > 64 || len(name) < 1 {
		return false
	}
	return validTopicChannelNameRegex.MatchString(name)
}
