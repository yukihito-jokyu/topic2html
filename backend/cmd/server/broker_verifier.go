//go:build !e2e

package main

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"

	"github.com/yukihito-jokyu/topic2html/backend/repository/codex"
)

var (
	dialBroker    = (&net.Dialer{}).DialContext
	socketOwnerID = func(info os.FileInfo) uint32 { return info.Sys().(*syscall.Stat_t).Uid }
	serverUserID  = os.Geteuid
)

func brokerVerifier() func(context.Context, string) error {
	return verifyBrokerEndpoint
}

func verifyBrokerEndpoint(ctx context.Context, endpoint string) error {
	broker, err := codex.NewClient(endpoint)
	if err != nil {
		return err
	}
	info, err := os.Stat(broker.Endpoint())
	if err != nil || info.Mode()&os.ModeSocket == 0 || info.Mode().Perm() != 0660 {
		return errors.New("broker endpoint permissions are unsafe")
	}
	ownerID := socketOwnerID(info)
	if int(ownerID) == serverUserID() {
		return errors.New("broker endpoint owner is unsafe")
	}
	connection, err := dialBroker(ctx, "unix", broker.Endpoint())
	if err != nil {
		return err
	}

	return connection.Close()
}
