package database

import "testing"

func TestResolveAttachmentMime(t *testing.T) {
	mp4 := append([]byte{0x00, 0x00, 0x00, 0x18}, []byte("ftypisom0000")...)
	webm := append([]byte{0x1A, 0x45, 0xDF, 0xA3}, make([]byte, 16)...)
	png := append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 16)...)
	zip := append([]byte("PK\x03\x04"), make([]byte, 16)...)

	cases := []struct {
		name string
		data []byte
		mime string
		kind string
	}{
		{"mp4", mp4, "video/mp4", "video"},
		{"webm", webm, "video/webm", "video"},
		{"png", png, "image/png", "image"},
		{"zip", zip, "application/zip", "file"},
		{"plain text", []byte("just some plain notes\n"), "text/plain", "document"},
		// Markup must never resolve to a renderable type — markup sniffed as
		// text/html is not whitelisted, and verifyTextLike rejects it under
		// text/plain, so it falls through to download-only octet-stream.
		{"html", []byte("<!DOCTYPE html><script>alert(1)</script>"), "application/octet-stream", "file"},
		{"svg as text", []byte("<svg xmlns='http://www.w3.org/2000/svg'></svg>"), "application/octet-stream", "file"},
		{"random binary", []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}, "application/octet-stream", "file"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mime := ResolveAttachmentMime(tc.data)
			if mime != tc.mime {
				t.Fatalf("ResolveAttachmentMime = %q, want %q", mime, tc.mime)
			}
			if kind := AttachmentKind(mime); kind != tc.kind {
				t.Fatalf("AttachmentKind(%q) = %q, want %q", mime, kind, tc.kind)
			}
		})
	}
}

func TestVerifyAttachmentMagicRejectsRenamedBytes(t *testing.T) {
	html := []byte("<!DOCTYPE html><p>not a video</p>")
	for _, mime := range []string{"video/mp4", "video/webm", "image/png", "application/zip"} {
		if VerifyAttachmentMagic(mime, html) {
			t.Errorf("VerifyAttachmentMagic(%q) accepted HTML bytes", mime)
		}
	}
	mp4 := append([]byte{0x00, 0x00, 0x00, 0x18}, []byte("ftypisom0000")...)
	if !VerifyAttachmentMagic("video/mp4", mp4) {
		t.Errorf("VerifyAttachmentMagic rejected a genuine mp4 header")
	}
}

func TestAttachmentURLSuffix(t *testing.T) {
	cases := map[string]string{
		"video/mp4":                ".mp4",
		"video/webm":               ".webm",
		"image/png":                "",
		"application/pdf":          "",
		"application/octet-stream": "",
	}
	for mime, want := range cases {
		if got := AttachmentURLSuffix(mime); got != want {
			t.Errorf("AttachmentURLSuffix(%q) = %q, want %q", mime, got, want)
		}
	}
}
