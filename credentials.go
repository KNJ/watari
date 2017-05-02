package watari

import (
	"encoding/json"
	"os"
)

// Credentials ...
type Credentials struct {
	Username string
	Password string
}

// Load sets self credentials data.
func (cred *Credentials) Load(filePath string) error {
	// Open credential JSON file
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return err
	}

	// Decode JSON
	dec := json.NewDecoder(file)
	err = dec.Decode(cred)

	return err
}
