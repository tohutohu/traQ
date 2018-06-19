package router

import (
	"bytes"
	"context"
	"github.com/satori/go.uuid"
	"github.com/traPtitech/traQ/event"
	"github.com/traPtitech/traQ/external/imagemagick"
	"github.com/traPtitech/traQ/model"
	"github.com/traPtitech/traQ/oauth2"
	"github.com/traPtitech/traQ/utils/thumb"
	"image"
	_ "image/jpeg" // image.Decode用
	_ "image/png"  // image.Decode用
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

var errMySQLDuplicatedRecord uint16 = 1062

// Handlers ハンドラ
type Handlers struct {
	Bot    *event.BotProcessor
	OAuth2 *oauth2.Handler
}

// CustomHTTPErrorHandler :json形式でエラーレスポンスを返す
func CustomHTTPErrorHandler(err error, c echo.Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = map[string]interface{}{"message": msg}
	}

	if err = c.JSON(code, msg); err != nil {
		c.Echo().Logger.Errorf("an error occurred while sending to JSON: %v", err)
	}

}

func bindAndValidate(c echo.Context, i interface{}) error {
	if err := c.Bind(i); err != nil {
		return err
	}
	if err := c.Validate(i); err != nil {
		return err
	}
	return nil
}

func processMultipartFormIconUpload(c echo.Context, file *multipart.FileHeader) (uuid.UUID, error) {
	// ファイルサイズ制限1MB
	if file.Size > 1024*1024 {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "too large image file (1MB limit)")
	}

	// ファイルタイプ確認・必要があればリサイズ
	src, err := file.Open()
	if err != nil {
		c.Logger().Error(err)
		return uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	b, err := processIcon(c, file.Header.Get(echo.HeaderContentType), src)
	src.Close()
	if err != nil {
		return uuid.Nil, err
	}

	// アイコン画像保存
	fileID, err := saveFile(file.Filename, b)
	if err != nil {
		c.Logger().Error(err)
		return uuid.Nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	return fileID, nil
}

func saveFile(name string, src *bytes.Buffer) (uuid.UUID, error) {
	file := &model.File{
		Name: name,
		Size: int64(src.Len()),
	}
	if err := file.Create(src); err != nil {
		return uuid.Nil, err
	}

	return uuid.Must(uuid.FromString(file.ID)), nil
}

func processIcon(c echo.Context, mime string, src io.Reader) (*bytes.Buffer, error) {
	switch mime {
	case "image/png", "image/jpeg":
		return processStillIconImage(c, src)
	case "image/gif":
		return processGifIconImage(c, src)
	}
	return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid image file")
}

func processStillIconImage(c echo.Context, src io.Reader) (*bytes.Buffer, error) {
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "bad image file")
	}

	if img.Bounds().Size().X > iconMaxWidth || img.Bounds().Size().Y > iconMaxHeight {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //10秒以内に終わらないファイルは無効
		defer cancel()
		img, err = thumb.Resize(ctx, img, iconMaxWidth, iconMaxHeight)
		if err != nil {
			switch err {
			case context.DeadlineExceeded:
				// リサイズタイムアウト
				return nil, echo.NewHTTPError(http.StatusBadRequest, "bad image file (resize timeout)")
			default:
				// 予期しないエラー
				c.Logger().Error(err)
				return nil, echo.NewHTTPError(http.StatusInternalServerError)
			}
		}
	}

	// bytesに戻す
	b, err := thumb.EncodeToPNG(img)
	if err != nil {
		// 予期しないエラー
		c.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}

	return b, nil
}

func processGifIconImage(c echo.Context, src io.Reader) (*bytes.Buffer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //10秒以内に終わらないファイルは無効
	defer cancel()

	b, err := imagemagick.ResizeAnimationGIF(ctx, src, iconMaxWidth, iconMaxHeight, false)
	if err != nil {
		switch err {
		case imagemagick.ErrUnavailable:
			// gifは一時的にサポートされていない
			return nil, echo.NewHTTPError(http.StatusBadRequest, "gif file is temporarily unsupported")
		case imagemagick.ErrUnsupportedType:
			// 不正なgifである
			return nil, echo.NewHTTPError(http.StatusBadRequest, "bad image file")
		case context.DeadlineExceeded:
			// リサイズタイムアウト
			return nil, echo.NewHTTPError(http.StatusBadRequest, "bad image file (resize timeout)")
		default:
			// 予期しないエラー
			c.Logger().Error(err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError)
		}
	}

	return b, nil
}
