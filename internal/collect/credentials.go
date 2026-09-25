package collect

import (
	"encoding/json"
	"fmt"
	"os"
)

// AdminKey reads the local credential only when no explicit environment value exists.
func AdminKey(path string) (string, error) {
	if key := os.Getenv("OPENAI_ADMIN_KEY"); key != "" {
		return key, nil
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode().Perm()&0077 != 0 {
		return "", fmt.Errorf("credential file permissions must be 0600")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var credentials struct {
		OpenAIAdminKey string `json:"openai_admin_key"`
	}
	if err := json.Unmarshal(data, &credentials); err != nil {
		return "", err
	}
	return credentials.OpenAIAdminKey, nil
}
