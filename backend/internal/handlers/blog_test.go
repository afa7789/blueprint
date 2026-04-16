package handlers_test

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/testutil"
	"github.com/afa/blueprint/backend/pkg/config"
)

func setupBlogApp() (*fiber.App, *testutil.MockBlogRepo) {
	app := fiber.New()
	blogRepo := testutil.NewMockBlogRepo()
	cfg := &config.Config{
		Env:         "development",
		FrontendURL: "http://localhost:3000",
	}

	h := handlers.NewBlogHandler(blogRepo, cfg)

	app.Get("/blog", h.ListPublished)
	app.Get("/blog/rss.xml", h.RSSFeed)
	app.Get("/blog/atom.xml", h.AtomFeed)
	app.Get("/blog/:slug", h.GetBySlug)

	return app, blogRepo
}

func TestRSSFeed_Empty(t *testing.T) {
	app, _ := setupBlogApp()

	req := httptest.NewRequest(http.MethodGet, "/blog/rss.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/rss+xml; charset=utf-8" {
		t.Fatalf("expected application/rss+xml, got %s", ct)
	}

	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	xmlStr := string(body[:n])

	if !contains(xmlStr, "<rss") {
		t.Fatal("RSS feed missing <rss> tag")
	}
	if !contains(xmlStr, "version=\"2.0\"") {
		t.Fatal("RSS feed missing version attribute")
	}
	if !contains(xmlStr, "<channel>") {
		t.Fatal("RSS feed missing <channel> tag")
	}
}

func TestRSSFeed_WithPosts(t *testing.T) {
	app, blogRepo := setupBlogApp()

	pubTime := time.Now()
	excerpt := "This is a test excerpt"
	blogRepo.Create(nil, &domain.BlogPost{
		ID:          "post-1",
		Title:       "Test Post 1",
		Slug:        "test-post-1",
		Excerpt:     &excerpt,
		Status:      "published",
		PublishedAt: &pubTime,
		CreatedAt:   time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/blog/rss.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	xmlStr := string(body[:n])

	if !contains(xmlStr, "Test Post 1") {
		t.Fatal("RSS feed missing post title")
	}
	if !contains(xmlStr, "test-post-1") {
		t.Fatal("RSS feed missing post slug in link")
	}
	if !contains(xmlStr, "This is a test excerpt") {
		t.Fatal("RSS feed missing post excerpt")
	}
	if !contains(xmlStr, "<item>") {
		t.Fatal("RSS feed missing <item> tag")
	}
}

func TestAtomFeed_Empty(t *testing.T) {
	app, _ := setupBlogApp()

	req := httptest.NewRequest(http.MethodGet, "/blog/atom.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/atom+xml; charset=utf-8" {
		t.Fatalf("expected application/atom+xml, got %s", ct)
	}

	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	xmlStr := string(body[:n])

	if !contains(xmlStr, "<feed") {
		t.Fatal("Atom feed missing <feed> tag")
	}
	if !contains(xmlStr, "xmlns=\"http://www.w3.org/2005/Atom\"") {
		t.Fatal("Atom feed missing xmlns attribute")
	}
}

func TestAtomFeed_WithPosts(t *testing.T) {
	app, blogRepo := setupBlogApp()

	pubTime := time.Now()
	excerpt := "This is a test excerpt"
	blogRepo.Create(nil, &domain.BlogPost{
		ID:          "post-1",
		Title:       "Test Post 1",
		Slug:        "test-post-1",
		Excerpt:     &excerpt,
		Status:      "published",
		PublishedAt: &pubTime,
		CreatedAt:   time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/blog/atom.xml", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	xmlStr := string(body[:n])

	if !contains(xmlStr, "Test Post 1") {
		t.Fatal("Atom feed missing post title")
	}
	if !contains(xmlStr, "test-post-1") {
		t.Fatal("Atom feed missing post slug in ID/link")
	}
	if !contains(xmlStr, "This is a test excerpt") {
		t.Fatal("Atom feed missing post excerpt as summary")
	}
	if !contains(xmlStr, "<entry>") {
		t.Fatal("Atom feed missing <entry> tag")
	}
}

func TestFeed_ValidXML(t *testing.T) {
	app, blogRepo := setupBlogApp()

	pubTime := time.Now()
	excerpt := "Test excerpt"
	blogRepo.Create(nil, &domain.BlogPost{
		ID:          "post-1",
		Title:       "Test Post",
		Slug:        "test-post",
		Excerpt:     &excerpt,
		Status:      "published",
		PublishedAt: &pubTime,
		CreatedAt:   time.Now(),
	})

	// Test RSS XML validity
	req := httptest.NewRequest(http.MethodGet, "/blog/rss.xml", nil)
	resp, _ := app.Test(req)
	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	resp.Body.Close()

	var rss interface{}
	if err := xml.Unmarshal(body[:n], &rss); err != nil {
		t.Fatalf("RSS feed is not valid XML: %v", err)
	}

	// Test Atom XML validity
	req = httptest.NewRequest(http.MethodGet, "/blog/atom.xml", nil)
	resp, _ = app.Test(req)
	body = make([]byte, 10000)
	n, _ = resp.Body.Read(body)
	resp.Body.Close()

	if err := xml.Unmarshal(body[:n], &rss); err != nil {
		t.Fatalf("Atom feed is not valid XML: %v", err)
	}
}

func TestFeed_OnlyPublishedPosts(t *testing.T) {
	app, blogRepo := setupBlogApp()

	pubTime := time.Now()
	excerpt := "Published"
	blogRepo.Create(nil, &domain.BlogPost{
		ID:          "post-1",
		Title:       "Published Post",
		Slug:        "published-post",
		Excerpt:     &excerpt,
		Status:      "published",
		PublishedAt: &pubTime,
		CreatedAt:   time.Now(),
	})

	draftExcerpt := "Draft"
	blogRepo.Create(nil, &domain.BlogPost{
		ID:       "post-2",
		Title:    "Draft Post",
		Slug:     "draft-post",
		Excerpt:  &draftExcerpt,
		Status:   "draft",
		CreatedAt: time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/blog/rss.xml", nil)
	resp, _ := app.Test(req)
	body := make([]byte, 10000)
	n, _ := resp.Body.Read(body)
	resp.Body.Close()
	xmlStr := string(body[:n])

	if !contains(xmlStr, "Published Post") {
		t.Fatal("Published post missing from feed")
	}
	if contains(xmlStr, "Draft Post") {
		t.Fatal("Draft post should not be in feed")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
