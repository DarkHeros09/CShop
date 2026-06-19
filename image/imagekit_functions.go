package image

import (
	"context"

	"github.com/imagekit-developer/imagekit-go/v2"
	"github.com/imagekit-developer/imagekit-go/v2/shared"
)

func (image *ImageKit) ListAndSearch(ctx context.Context, params imagekit.AssetListParams) (*[]imagekit.AssetListResponseUnion, error) {
	resp, err := image.ik.Assets.List(ctx, params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (image *ImageKit) UrlGeneration(ctx context.Context, params shared.SrcOptionsParam) (*string, error) {
	url := image.ik.Helper.BuildURL(params)
	return &url, nil
}
