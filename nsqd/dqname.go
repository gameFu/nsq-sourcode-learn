package nsqd

func getBackendName(topicName, channelName string) string {
	backendName := topicName + ";" + channelName
	return backendName
}
