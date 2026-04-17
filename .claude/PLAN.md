# Blueprint - Feature Roadmap

## Blog Features

- ✅ **RSS 2.0 Feed** (2026-04-16)
  - Endpoint: `GET /api/v1/blog/rss.xml`
  - Last 20 published posts
  - Uses excerpt for description
  - Links to frontend blog pages
  
- ✅ **Atom 1.0 Feed** (2026-04-16)
  - Endpoint: `GET /api/v1/blog/atom.xml`
  - Last 20 published posts
  - Uses excerpt for description
  - Links to frontend blog pages
  
Both feeds are gated by the existing `blog_enabled` feature flag.
