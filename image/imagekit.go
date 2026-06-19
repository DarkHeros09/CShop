package image

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/option"
	"github.com/imagekit-developer/imagekit-go/v2/shared"
)

type ImageKitManagement interface {
	ListAndSearch(ctx context.Context, params imagekit.AssetListParams) (*[]imagekit.AssetListResponseUnion, error)
	UrlGeneration(ctx context.Context, params shared.SrcOptionsParam) (*string, error)
}

type ImageKit struct {
	ik *imagekit.Client
}

func NewImageKit(privateKey string) ImageKitManagement {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
	}
	ik := imagekit.NewClient(
		option.WithPrivateKey(privateKey),
		option.WithHTTPClient(client),
	)

	return &ImageKit{
		ik: &ik,
	}
}
