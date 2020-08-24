package main

import (
	"context"

	"github.com/wwq-2020/ingress/pkg/server"
	"github.com/wwq-2020/ingress/pkg/util"
	"github.com/wwq1988/group"
)

func main() {
	clientSet := util.MustGetClientSet()

	server := server.New(clientSet)
	group.Go(func(ctx context.Context) {
		server.Start()
	})
	group.AddShutdownHook(server.Stop)
	group.Wait()

}
