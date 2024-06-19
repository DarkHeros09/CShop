package image

import (
	"context"

	"github.com/imagekit-developer/imagekit-go"
	"github.com/imagekit-developer/imagekit-go/api/media"
)

type ImageKitManagement interface {
	ListAndSearch(ctx context.Context, params media.FilesParam) (*media.FilesResponse, error)
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
