package specs

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestYAMLFilesParse(t *testing.T) {
	for _, path := range []string{
		"openapi.yaml",
		"../docker-compose.yml",
	} {
		t.Run(path, func(t *testing.T) {
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			var value any
			if err := yaml.Unmarshal(content, &value); err != nil {
				t.Fatalf("yaml.Unmarshal() error = %v", err)
			}
		})
	}
}
