package store

// EscapeString escapes single quotes. Used to create Postgres-compatible strings
func EscapeString(input string) (output string) {
	for _, c := range input {
		if c == '\'' {
			output += "'"
		}
		output += string(c)
	}
	return
}
