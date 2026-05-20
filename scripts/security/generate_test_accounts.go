//go:build ignore

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type TestAccount struct {
	UserID      string    `json:"user_id"`
	AccountID   string    `json:"account_id"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Tier        int       `json:"tier"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

func main() {
	accounts := []TestAccount{}

	accounts = append(accounts, TestAccount{
		UserID:      "admin_001",
		AccountID:   "admin_pen_test",
		PhoneNumber: "13900000001",
		Email:       "admin.pentest@example.com",
		Password:    "AdminTest123!",
		Tier:        4,
		Role:        "admin",
		CreatedAt:   time.Now(),
	})

	tiers := []int{0, 2, 3, 4}
	for i := 0; i < 8; i++ {
		salt := generateRandomHex(4)
		accounts = append(accounts, TestAccount{
			UserID:      fmt.Sprintf("pentest_%s_%03d", salt, i+1),
			AccountID:   fmt.Sprintf("pentest_user_%s_%03d", salt, i+1),
			PhoneNumber: fmt.Sprintf("138%08d", i+100),
			Email:       fmt.Sprintf("pentest_%s_%03d@example.com", salt, i+1),
			Password:    fmt.Sprintf("TestPass%d!%s", i+1, salt[:4]),
			Tier:        tiers[i%len(tiers)],
			Role:        "normal",
			CreatedAt:   time.Now(),
		})
	}

	for _, payload := range []string{"' OR '1'='1", "'; DROP TABLE users;--", "1 UNION SELECT * FROM users"} {
		salt := generateRandomHex(4)
		accounts = append(accounts, TestAccount{
			UserID:      fmt.Sprintf("sqli_%s", salt),
			AccountID:   fmt.Sprintf("sqli_%s", salt),
			PhoneNumber: payload,
			Email:       payload + "@test.com",
			Password:    "SqliTest123!",
			Tier:        0,
			Role:        "sqli_test",
			CreatedAt:   time.Now(),
		})
	}

	for _, payload := range []string{"<script>alert(1)</script>", "<img src=x onerror=alert(1)>"} {
		salt := generateRandomHex(4)
		accounts = append(accounts, TestAccount{
			UserID:      fmt.Sprintf("xss_%s", salt),
			AccountID:   fmt.Sprintf("xss_%s", salt),
			PhoneNumber: fmt.Sprintf("137%08d", time.Now().UnixNano()%100000000),
			Email:       payload,
			Password:    "XssTest123!",
			Tier:        0,
			Role:        "xss_test",
			CreatedAt:   time.Now(),
		})
	}

	data, _ := json.MarshalIndent(accounts, "", "  ")
	fmt.Println(string(data))

	os.MkdirAll("reports/security", 0755)
	os.WriteFile("reports/security/test_accounts.json", data, 0644)
	fmt.Fprintf(os.Stderr, "\nGenerated %d test accounts → reports/security/test_accounts.json\n", len(accounts))
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
