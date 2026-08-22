//go:build e2e

package main

import "context"

func brokerVerifier() func(context.Context, string) error {
	return func(context.Context, string) error { return nil }
}
