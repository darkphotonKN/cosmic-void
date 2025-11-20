package discovery

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ServiceConnection(ctx context.Context, serviceName string, registry Registry) (*grpc.ClientConn, error) {

	fmt.Println("serviceName: ", serviceName)
	// discover the other services
	addrs, err := registry.Discover(ctx, serviceName)

	fmt.Println("addrs: ", addrs)
	if err != nil {
		return nil, err
	}

	length := len(addrs)

	if length == 0 {
		return nil, errors.New("There are no services to discover now.")
	}
	// credentials := insecure.NewCredentials()

	return grpc.NewClient(addrs[rand.Intn(length)], grpc.WithTransportCredentials(insecure.NewCredentials()))
}
