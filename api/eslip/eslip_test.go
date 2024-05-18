package eslip

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUploadToS3(t *testing.T) {
	t.Run("should able to upload image to S3 bucket", func(t *testing.T) {
		f, _ := os.CreateTemp("", "eslip1.jpg")
		defer f.Close()
		defer os.Remove(f.Name())

		e := echo.New()
		defer e.Close()

		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		loc, err := UploadToS3(c, "eslip1.jpg", f)

		assert.NoError(t, err)
		assert.Equal(t, "location/on/s3/bucket/eslip1.jpg", loc)
	})

}

func TestUpload(t *testing.T) {
	t.Run("should able to upload image", func(t *testing.T) {
		e := echo.New()
		defer e.Close()

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("images", "test.jpg")
		assert.NoError(t, err)
		_, err = io.Copy(part, strings.NewReader("fake image content"))
		assert.NoError(t, err)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		Upload(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Image uploaded successfully")
	})
}
