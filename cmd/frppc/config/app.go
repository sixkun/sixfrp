package config

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/goravel/framework/support/carbon"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"
)

// Boot loads configuration after command-line environment overrides are set.
func Boot() {
	loadApp()
	loadFrppc()
	loadLogging()
}

// generateAppKey creates a 32-character key from FRPPC_ID and FRPPC_SECRET
// The key contains only alphanumeric characters (a-zA-Z0-9)
func generateAppKey(id, secret string) string {
	combined := id + ":" + secret
	hash := sha256.Sum256([]byte(combined))

	// Use base64 encoding and filter to alphanumeric only
	encoded := base64.RawStdEncoding.EncodeToString(hash[:])

	// Filter to alphanumeric characters only
	result := make([]byte, 0, 32)
	for i := 0; i < len(encoded) && len(result) < 32; i++ {
		c := encoded[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result = append(result, c)
		}
	}

	return string(result)
}

func loadApp() {
	config := facades.Config()

	// Generate APP_KEY from FRPPC_ID and FRPPC_SECRET
	frppcID := config.Env("FRPPC_ID", "").(string)
	frppcSecret := config.Env("FRPPC_SECRET", "").(string)

	if frppcID == "" || frppcSecret == "" {
		panic("FRPPC_ID and FRPPC_SECRET must be set")
	}

	appKey := generateAppKey(frppcID, frppcSecret)

	config.Add("app", map[string]any{
		// Application Name
		//
		// This value is the name of your application. This value is used when the
		// framework needs to place the application's name in a notification or
		// any other location as required by the application or its packages.
		"name": config.Env("FRPPC_APP_NAME", "Goravel"),
		// Application Environment
		//
		// This value determines the "environment" your application is currently
		// running in. This may determine how you prefer to configure various
		// services the application utilizes. Set this in your ".env" file.
		"env": config.Env("FRPPC_APP_ENV", "production"),

		// Application Debug Mode
		"debug": config.Env("FRPPC_APP_DEBUG", false),

		// Application Timezone
		//
		// Here you may specify the default timezone for your application.
		// Example: UTC, Asia/Shanghai
		// More: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
		"timezone": carbon.UTC,

		// Application Locale Configuration
		//
		// The application locale determines the default locale that will be used
		// by the translation service provider. You are free to set this value
		// to any of the locales which will be supported by the application.
		"locale": "en",

		// Application Fallback Locale
		//
		// The fallback locale determines the locale to use when the current one
		// is not available. You may change the value to correspond to any of
		// the language folders that are provided through your application.
		"fallback_locale": "en",

		// Encryption Key
		//
		// 32 character string, otherwise these encrypted strings
		// will not be safe. Please do this before deploying an application!
		"key": appKey,
	})
}
