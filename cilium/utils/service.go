package utils

// LookupServiceName returns the service name from some pre defined labels.
// Can be from "com.docker.compose.service" and "com.intent.service".
func LookupServiceName(labels map[string]string) string {

	if serviceName, ok := labels["com.intent.service"]; ok {
		return serviceName
	}

	if serviceName, ok := labels["com.docker.compose.service"]; ok {
		return serviceName
	}

	return ""
}
