package image

import (
	"context"

	"github.com/imagekit-developer/imagekit-go/api/media"
	ikurl "github.com/imagekit-developer/imagekit-go/url"
)

func (image *ImageKit) ListAndSearch(ctx context.Context, params media.FilesParam) (*media.FilesResponse, error) {
	resp, err := image.ik.Media.Files(ctx, params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (image *ImageKit) UrlGeneration(ctx context.Context, params ikurl.UrlParam) (*string, error) {
	url, err := image.ik.Url(params)
	if err != nil {
		return nil, err
	}
	return &url, nil
}
