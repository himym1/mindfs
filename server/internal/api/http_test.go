package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPathForStaticAssetCleansURLPaths(t *testing.T) {
	tests := []struct {
		name        string
		requestPath string
		want        string
	}{
		{
			name:        "absolute asset path",
			requestPath: "/assets/app.js",
			want:        "assets/app.js",
		},
		{
			name:        "duplicate slash path",
			requestPath: "//assets/app.js",
			want:        "assets/app.js",
		},
		{
			name:        "root path",
			requestPath: "/",
			want:        "",
		},
		{
			name:        "relayed asset alias",
			requestPath: "/mindfs-assets/index.js",
			want:        "assets/index.js",
		},
		{
			name:        "relay node asset prefix",
			requestPath: "/n/ZroV4sfU/assets/index.js",
			want:        "assets/index.js",
		},
		{
			name:        "relay node mindfs asset prefix",
			requestPath: "/n/ZroV4sfU/mindfs-assets/index.js",
			want:        "assets/index.js",
		},
		{
			name:        "relay node service worker prefix",
			requestPath: "/n/ZroV4sfU/service-worker.js",
			want:        "service-worker.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathForStaticAsset(tt.requestPath)
			if got != tt.want {
				t.Fatalf("pathForStaticAsset(%q) = %q, want %q", tt.requestPath, got, tt.want)
			}
		})
	}
}

func TestRewriteRelayedFrontendContentKeepsAssetsUnderRelayNode(t *testing.T) {
	input := `<script src="./assets/index.js"></script><link href="./assets/index.css">`
	got := rewriteRelayedFrontendContent(input)
	if strings.Contains(got, `"/mindfs-assets/`) {
		t.Fatalf("rewrite used absolute relay asset path: %s", got)
	}
	if !strings.Contains(got, `"./mindfs-assets/index.js"`) || !strings.Contains(got, `"./mindfs-assets/index.css"`) {
		t.Fatalf("rewrite did not keep relative relay asset paths: %s", got)
	}
}

func TestRewriteRelayedFrontendContentKeepsInstallManifestUnderRelayNode(t *testing.T) {
	input := `<link rel="manifest" href="/manifest.webmanifest"><link rel="apple-touch-icon" href="/apple-touch-icon.png">`
	got := rewriteRelayedFrontendContent(input)
	if strings.Contains(got, `href="/manifest.webmanifest"`) || strings.Contains(got, `href="/apple-touch-icon.png"`) {
		t.Fatalf("rewrite left install resources outside relay node scope: %s", got)
	}
	if !strings.Contains(got, `href="./manifest.webmanifest"`) || !strings.Contains(got, `href="./apple-touch-icon.png"`) {
		t.Fatalf("rewrite did not keep install resources relative to relay node: %s", got)
	}
}

func TestRelayedServiceWorkerRemainsInstallable(t *testing.T) {
	if strings.Contains(relayedServiceWorkerScript, "unregister") {
		t.Fatalf("relayed service worker unregisters itself, breaking PWA installability: %s", relayedServiceWorkerScript)
	}
	if !strings.Contains(relayedServiceWorkerScript, `self.addEventListener("fetch"`) {
		t.Fatalf("relayed service worker must install a fetch listener for PWA installability: %s", relayedServiceWorkerScript)
	}
}

func TestFallbackFrontendEntryAssetUsesCurrentIndexReference(t *testing.T) {
	staticDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(staticDir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	index := `<!doctype html><script type="module" src="./assets/index-current.js"></script><link rel="stylesheet" href="./assets/index-current.css">`
	if err := os.WriteFile(filepath.Join(staticDir, "index.html"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "assets", "index-current.js"), []byte("console.log('current')"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staticDir, "assets", "index-current.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := fallbackFrontendEntryAsset(staticDir, "assets/index-stale.js"); got != "assets/index-current.js" {
		t.Fatalf("fallback js = %q, want assets/index-current.js", got)
	}
	if got := fallbackFrontendEntryAsset(staticDir, "assets/index-stale.css"); got != "assets/index-current.css" {
		t.Fatalf("fallback css = %q, want assets/index-current.css", got)
	}
	if got := fallbackFrontendEntryAsset(staticDir, "assets/other-stale.js"); got != "" {
		t.Fatalf("non-entry fallback = %q, want empty", got)
	}
}
