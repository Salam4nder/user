package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
)

// Metadata is a struct that contains metadata about the request.
type Metadata struct {
	UserAgent string
	ClientIP  string
}

// MetadataFromContext returns metadata from the context.
func MetadataFromContext(ctx context.Context) *Metadata {
	mtdt := new(Metadata)

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtdt.UserAgent = userAgents[0]
		}

		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtdt.ClientIP = clientIPs[0]
		}
	}

	if peer, ok := peer.FromContext(ctx); ok {
		mtdt.ClientIP = peer.Addr.String()
	}

	return mtdt
}
