package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	apiv1 "github.com/slok/simple-ingress-external-auth/pkg/api/v1"
)

func run(ctx context.Context) error {
	const tokenQ = 1000000

	tokens := []apiv1.Token{}

	for i := 0; i < tokenQ; i++ {
		token, err := genToken()
		if err != nil {
			return err
		}
		tokens = append(tokens, apiv1.Token{Value: token})
	}
	c := apiv1.Config{
		Version: "v1",
		Tokens:  tokens,
	}

	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	_, err = os.Stdout.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func genToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "App error: %s", err)
		os.Exit(1)
		return
	}
}
