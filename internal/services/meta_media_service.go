package services

import (
	"fmt"

	"projectvows/internal/config"
)

// MetaMediaService abstracts uploading binary media to the Meta WhatsApp
// Media API so the returned media id can be reused in messages.
type MetaMediaService interface {
	// UploadImage uploads PNG bytes and returns the Meta media id.
	UploadImage(data []byte, filename string) (mediaID string, err error)
}

type metaMediaServiceStub struct {
	cfg *config.Config
}

// NewMetaMediaService returns the stub implementation.
func NewMetaMediaService(cfg *config.Config) MetaMediaService {
	return &metaMediaServiceStub{cfg: cfg}
}

func (s *metaMediaServiceStub) UploadImage(data []byte, filename string) (string, error) {
	// TODO: Replace with a real multipart upload to the Meta Media API:
	//   POST https://graph.facebook.com/{API_VERSION}/{PHONE_NUMBER_ID}/media
	//   Authorization: Bearer {META_ACCESS_TOKEN}
	//   form fields: messaging_product=whatsapp, type=image/png, file=<bytes>
	// The real response contains {"id": "<media_id>"}.
	return fmt.Sprintf("STUB-MEDIA-%s-%d", filename, len(data)), nil
}
