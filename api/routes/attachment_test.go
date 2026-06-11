package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func testContext(t *testing.T, req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	return c, rec
}

func TestAttachmentKeyParam(t *testing.T) {
	cases := map[string]string{
		"4f1c2d3e-aaaa-bbbb-cccc-000000000000":      "4f1c2d3e-aaaa-bbbb-cccc-000000000000",
		"4f1c2d3e-aaaa-bbbb-cccc-000000000000.mp4":  "4f1c2d3e-aaaa-bbbb-cccc-000000000000",
		"4f1c2d3e-aaaa-bbbb-cccc-000000000000.webm": "4f1c2d3e-aaaa-bbbb-cccc-000000000000",
		".mp4": ".mp4", // leading dot is not a suffix; leave untouched
	}
	for param, want := range cases {
		c, _ := testContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
		c.Params = gin.Params{{Key: "key", Value: param}}
		if got := attachmentKeyParam(c); got != want {
			t.Errorf("attachmentKeyParam(%q) = %q, want %q", param, got, want)
		}
	}
}

func TestServeBytesWithRangesFullResponse(t *testing.T) {
	data := []byte("0123456789")
	c, rec := testContext(t, httptest.NewRequest(http.MethodGet, "/", nil))

	serveBytesWithRanges(c, "video/mp4", data, "clip.mp4")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "0123456789" {
		t.Errorf("body = %q", got)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "video/mp4" {
		t.Errorf("Content-Type = %q", ct)
	}
	if rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Errorf("Accept-Ranges missing; video seeking needs it")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("nosniff header missing")
	}
	if rec.Header().Get("ETag") == "" {
		t.Errorf("ETag missing")
	}
	if cd := rec.Header().Get("Content-Disposition"); cd != `inline; filename="clip.mp4"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

func TestServeBytesWithRangesPartialContent(t *testing.T) {
	data := []byte("0123456789")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Range", "bytes=2-5")
	c, rec := testContext(t, req)

	serveBytesWithRanges(c, "video/mp4", data, "")

	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206", rec.Code)
	}
	if got := rec.Body.String(); got != "2345" {
		t.Errorf("body = %q, want %q", got, "2345")
	}
	if cr := rec.Header().Get("Content-Range"); cr != "bytes 2-5/10" {
		t.Errorf("Content-Range = %q", cr)
	}
}

func TestServeBytesWithRangesNotModified(t *testing.T) {
	data := []byte("0123456789")

	// First request to learn the ETag.
	c, rec := testContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
	serveBytesWithRanges(c, "video/mp4", data, "")
	etag := rec.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on first response")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", etag)
	c2, rec2 := testContext(t, req)
	serveBytesWithRanges(c2, "video/mp4", data, "")
	// gin buffers the status until the first body write; a 304 has no body,
	// so flush explicitly (the engine does this after handlers in production).
	c2.Writer.WriteHeaderNow()

	if rec2.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", rec2.Code)
	}
	if rec2.Body.Len() != 0 {
		t.Errorf("304 response carried a body (%d bytes)", rec2.Body.Len())
	}
}
