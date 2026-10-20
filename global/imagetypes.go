package global

import "strings"

// ImageExtensions 는 업로드·정적 서빙을 허용하는 이미지 확장자와, 그 확장자에
// 대해 http.DetectContentType 이 돌려줘야 하는 MIME 타입이다.
// SVG 는 스크립트를 담을 수 있어 제외한다.
var ImageExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// CanonicalImageExtension 은 스니핑한 MIME 타입에 대해 저장 시 쓸 확장자를 돌려준다.
var CanonicalImageExtension = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// IsImageExtension 은 확장자(대소문자 무시)가 허용 목록에 있는지 본다.
func IsImageExtension(ext string) bool {
	_, ok := ImageExtensions[strings.ToLower(ext)]
	return ok
}
