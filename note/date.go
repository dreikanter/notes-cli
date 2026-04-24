package note

// DateFormat is the canonical YYYYMMDD layout for UID-derived and CLI-facing
// dates. Use with time.Parse / time.Format (and ParseInLocation for UTC).
const DateFormat = "20060102"
