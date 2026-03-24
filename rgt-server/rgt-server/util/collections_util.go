package util

func MapToList(args map[string]string) []string {
	argsList := make([]string, 0, len(args))
	for key, value := range args {
		argsList = append(argsList, key+"="+value)
	}
	return argsList
}
