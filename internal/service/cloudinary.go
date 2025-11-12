package service

import (
	"context"
	"konnect/internal/logger"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"go.uber.org/zap"
)

type CloudinaryService struct {
	cld    *cloudinary.Cloudinary
	logger *zap.Logger
}

func NewCloudinaryService(logger *logger.Logger) (*CloudinaryService, error) {
	cld, err := cloudinary.New()
	if err != nil {
		return nil, err
	}
	cld.Config.URL.Secure = true

	return &CloudinaryService{
		cld:    cld,
		logger: logger.With(zap.String("component", "cloudinary_service")),
	}, nil
}

// UploadImage uploads an image to a cloudinary bucket. If the filename is provided, the asset is overwritten
func (s *CloudinaryService) UploadImage(ctx context.Context, file interface{}, folder, filename string) (url, publicID string, err error) {
	params := uploader.UploadParams{
		PublicID:       filename,
		Folder:         folder,
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(true),
	}
	// if a filename is provided, the file is replaced
	if filename != "" {
		params.UniqueFilename = api.Bool(false)
	}

	resp, err := s.cld.Upload.Upload(ctx, file, params)
	if err != nil {
		s.logger.Error("failed to upload image", zap.String("filename", filename), zap.Error(err))
		return "", "", err
	}

	s.logger.Info("successfully uploaded image",
		zap.String("publicID", resp.PublicID))

	return resp.SecureURL, resp.PublicID, nil
}

func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	if _, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID}); err != nil {
		s.logger.Error("failed to delete image", zap.String("publicID", publicID), zap.Error(err))
		return err
	}

	s.logger.Info("successfully deleted image", zap.String("publicId", publicID))
	return nil
}
