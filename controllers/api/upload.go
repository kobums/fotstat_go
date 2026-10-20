package api

import (
	"errors"
	"fmt"
	"fotstat/controllers"
	"fotstat/global"
	"fotstat/global/config"
	"fotstat/global/log"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type UploadController struct {
	controllers.Controller
}

// maxUploadSize 는 업로드 파일 1개의 최대 크기다.
const maxUploadSize = 20 * 1024 * 1024

// @POST()
// 이미지(jpg/png/gif/webp)만 받는다. 확장자 허용 목록과 내용 스니핑을 모두
// 통과해야 하며, 저장 파일명은 서버가 생성(날짜+UUID+정규 확장자)하므로
// 클라이언트가 보낸 파일명은 경로·확장자에 쓰이지 않는다.
func (c *UploadController) Index() {
	name := c.Context.FormValue("name", "upload")

	file, err := c.Context.FormFile(name)
	if err != nil {
		c.Error(errors.New("file is required"))
		return
	}

	if file.Size <= 0 || file.Size > maxUploadSize {
		c.Error(errors.New("invalid file size"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	expected, ok := global.ImageExtensions[ext]
	if !ok {
		c.Error(errors.New("only image files are allowed"))
		return
	}

	src, err := file.Open()
	if err != nil {
		c.Error(errors.New("cannot read file"))
		return
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	src.Close()
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		c.Error(errors.New("cannot read file"))
		return
	}
	sniffed := http.DetectContentType(head[:n])
	if sniffed != expected {
		c.Error(errors.New("only image files are allowed"))
		return
	}

	dirName := c.Now.DateAsOnlyNumber()
	fullPath := path.Join(config.UploadPath, dirName)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		err = os.Mkdir(fullPath, 0o755)
		if err != nil && !os.IsExist(err) {
			log.Error().Str("error", err.Error()).Msg("upload: mkdir")
			c.Error(errors.New("upload failed"))
			return
		}
	}

	filename := fmt.Sprintf("%v/%v_%v%v", dirName, dirName, strings.ReplaceAll(global.UUID(), "-", ""), global.CanonicalImageExtension[sniffed])
	if err := c.Context.SaveFile(file, path.Join(config.UploadPath, filename)); err != nil {
		log.Error().Str("error", err.Error()).Msg("upload: save")
		c.Error(errors.New("upload failed"))
		return
	}

	c.Set("filename", filename)
	c.Set("original", filepath.Base(file.Filename))
}
