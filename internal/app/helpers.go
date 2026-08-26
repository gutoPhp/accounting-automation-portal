package app

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var errNotFound = errors.New("registro não encontrado")

func newID(prefix string) string {
	buffer := make([]byte, 5)
	_, _ = rand.Read(buffer)
	return prefix + "-" + hex.EncodeToString(buffer)
}

func executionFinishedAt(start time.Time) time.Time {
	return start.Add(time.Since(start))
}

func sliceContains(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}

func readableError(err error) string {
	if errors.Is(err, errNotFound) {
		return "Registro não encontrado."
	}
	return fmt.Sprintf("Não foi possível concluir: %v", err)
}
