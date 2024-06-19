package image

import (
	"context"

	"github.com/imagekit-developer/imagekit-go/api/media"
)

func (image *ImageKit) ListAndSearch(ctx context.Context, params media.FilesParam) (*media.FilesResponse, error) {
	resp, err := image.ik.Media.Files(ctx, params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
