<template>
  <div class="blog-post-page">
    <router-link to="/blog" class="back-link">← Back to Blog</router-link>

    <div v-if="loading" class="loading">Loading post...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <article v-else-if="post" class="post-article">
      <div v-if="post.cover_image" class="post-cover">
        <img :src="post.cover_image" :alt="post.title" />
      </div>
      <header class="post-header">
        <p class="post-meta">
          <span v-if="post.author">{{ post.author }}</span>
          <span v-if="post.author && post.published_at"> · </span>
          <span v-if="post.published_at">{{ formatDate(post.published_at) }}</span>
        </p>
        <h1>{{ post.title }}</h1>
      </header>
      <div class="post-content" v-html="post.content"></div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '../../services/api'

interface Post {
  id: string
  title: string
  slug: string
  content: string
  cover_image: string | null
  author: string | null
  published_at: string | null
}

const route = useRoute()
const post = ref<Post | null>(null)
const loading = ref(false)
const error = ref('')

async function fetchPost() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ post: Post }>(`/api/v1/blog/${route.params.slug}`)
    post.value = data.post
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load post'
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

onMounted(fetchPost)
</script>

<style scoped>
.blog-post-page {
  padding: 32px;
  text-align: left;
  max-width: 760px;
  margin: 0 auto;
}

.back-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 14px;
  display: inline-block;
  margin-bottom: 24px;
}

.back-link:hover {
  text-decoration: underline;
}

.loading, .error {
  padding: 40px;
  text-align: center;
}

.error {
  color: #ef4444;
  border: 1px solid #ef4444;
  border-radius: 6px;
}

.post-cover {
  margin-bottom: 32px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border);
}

.post-cover img {
  width: 100%;
  height: auto;
  display: block;
}

.post-header {
  margin-bottom: 32px;
}

.post-meta {
  margin: 0 0 12px;
  font-size: 14px;
  color: var(--text);
}

.post-content {
  line-height: 1.7;
  color: var(--text);
}

.post-content :deep(h1),
.post-content :deep(h2),
.post-content :deep(h3) {
  color: var(--text-h);
  margin-top: 32px;
  margin-bottom: 12px;
}

.post-content :deep(p) {
  margin-bottom: 16px;
}

.post-content :deep(img) {
  max-width: 100%;
  border-radius: 6px;
}

.post-content :deep(a) {
  color: var(--accent);
}

.post-content :deep(pre) {
  background: var(--code-bg);
  padding: 16px;
  border-radius: 6px;
  overflow-x: auto;
}

@media (max-width: 600px) {
  .blog-post-page {
    padding: 20px;
  }
}
</style>
