package main

import (
	"flag"
	"fmt"
	"log"

	"course_management/config"
)

func main() {
	var environment string
	flag.StringVar(&environment, "env", "development", "Environment to generate secrets for (development, testing, staging, production)")
	flag.Parse()

	fmt.Printf("🔐 Generating secure secrets for %s environment\n", environment)

	// Validate environment
	validEnvs := []string{"development", "testing", "staging", "production"}
	valid := false
	for _, env := range validEnvs {
		if env == environment {
			valid = true
			break
		}
	}

	if !valid {
		log.Fatalf("❌ Invalid environment '%s'. Valid options: %v", environment, validEnvs)
	}

	// Generate secrets file
	if err := config.GenerateSecretsFile(environment); err != nil {
		log.Fatalf("❌ Failed to generate secrets file: %v", err)
	}

	fmt.Printf("✅ Secrets file generated successfully!\n")
	fmt.Printf("\n📝 Next steps:\n")
	fmt.Printf("1. Review the generated secrets file: config/%s.secrets.env\n", environment)
	fmt.Printf("2. Fill in the required values (database password, Google OAuth, etc.)\n")
	fmt.Printf("3. Keep this file secure and add it to .gitignore\n")
	fmt.Printf("4. Load the secrets when starting your application\n")

	if environment == "production" {
		fmt.Printf("\n⚠️  PRODUCTION SECURITY NOTES:\n")
		fmt.Printf("- Never commit secrets to version control\n")
		fmt.Printf("- Use environment variables or secret management services\n")
		fmt.Printf("- Rotate secrets regularly\n")
		fmt.Printf("- Monitor access to secret files\n")
	}

	// Show example usage
	fmt.Printf("\n💡 Example usage:\n")
	fmt.Printf("# Load environment and secrets:\n")
	fmt.Printf("export ENV=%s\n", environment)
	fmt.Printf("source config/%s.env\n", environment)
	fmt.Printf("source config/%s.secrets.env  # if exists\n", environment)
	fmt.Printf("./course_management\n")
}