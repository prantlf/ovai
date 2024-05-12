package log

func GetPlural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
