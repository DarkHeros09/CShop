package image

import (
	"context"

	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/media"
	ikurl "github.com/imagekit-developer/imagekit-go/url"
)

type ImageKitManagement interface {
	ListAndSearch(ctx context.Context, params media.FilesParam) (*media.FilesResponse, error)
	UrlGeneration(ctx context.Context, params ikurl.UrlParam) (*string, error)
}

type ImageKit struct {
	ik *imagekit.ImageKit
}

func NewImageKit(params imagekit.NewParams) ImageKitManagement {
	ik := imagekit.NewFromParams(params)
	return &ImageKit{
		ik: ik,
	}
}
