package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: hmac_signer <secret> <message>")
		os.Exit(1)
	}
	secret := os.Args[1]
	message := os.Args[2]
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signPayload := timestamp + ":" + message

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	fmt.Printf("X-Timestamp: %s\n", timestamp)
	fmt.Printf("X-Signature: %s\n", signature)
}
