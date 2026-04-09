<template>
  <div class="blog-index">
    <h1>Blog</h1>

    <div v-if="loading" class="loading">Loading posts...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="posts.length === 0" class="empty">No posts yet.</div>
    <div v-else class="posts-grid">
      <div
        v-for="post in posts"
        :key="post.id"
        class="post-card"
        @click="$router.push(`/blog/${post.slug}`)"
      >
        <div class="post-cover">
          <img v-if="post.cover_image" :src="post.cover_image" :alt="post.title" />
          <div v-else class="cover-placeholder"></div>
        </div>
        <div class="post-body">
          <p class="post-date">{{ formatDate(post.published_at || post.created_at) }}</p>
          <h2 class="post-title">{{ post.title }}</h2>
          <p class="post-excerpt">{{ post.excerpt }}</p>
        </div>
      </div>
    </div>

    <div v-if="totalPages > 1" class="pagination">
      <button :disabled="page === 1" @click="fetchPosts(page - 1)" class="btn">Previous</button>
      <span>Page {{ page }} of {{ totalPages }}</span>
      <button :disabled="page === totalPages" @click="fetchPosts(page + 1)" class="btn">Next</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'

interface Post {
  id: string
  title: string
  slug: string
  excerpt: string
  cover_image: string | null
  published_at: string | null
  created_at: string
}

interface PostsResponse {
  posts: Post[]
  total: number
  page: number
  limit: number
}

const posts = ref<Post[]>([])
const loading = ref(false)
const error = ref('')
const page = ref(1)
const totalPages = ref(1)
const limit = 10

async function fetchPosts(p = 1) {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<PostsResponse>(`/api/v1/blog?page=${p}&limit=${limit}`)
    posts.value = data.posts || []
    page.value = data.page || p
    totalPages.value = Math.ceil((data.total || 0) / limit)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load posts'
  } finally {
    loading.value = false
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })
}

onMounted(() => fetchPosts())
</script>

<style scoped>
.blog-index {
  padding: 32px;
  text-align: left;
  max-width: 900px;
  margin: 0 auto;
}

.loading, .empty {
  padding: 40px;
  text-align: center;
  color: var(--text);
}

.error {
  color: #ef4444;
  padding: 16px;
  border: 1px solid #ef4444;
  border-radius: 6px;
}

.posts-grid {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.post-card {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: 20px;
  border: 1px solid var(--border);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.post-card:hover {
  box-shadow: var(--shadow);
}

@media (max-width: 600px) {
  .post-card {
    grid-template-columns: 1fr;
  }
  .blog-index {
    padding: 20px;
  }
}

.post-cover {
  aspect-ratio: 4/3;
  overflow: hidden;
  background: var(--code-bg);
}

.post-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  background: var(--code-bg);
}

.post-body {
  padding: 20px;
}

.post-date {
  margin: 0 0 8px;
  font-size: 13px;
  color: var(--text);
}

.post-title {
  margin: 0 0 10px;
  font-size: 20px;
}

.post-excerpt {
  margin: 0;
  color: var(--text);
  font-size: 14px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 32px;
}

.btn {
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid var(--border);
  cursor: pointer;
  font-size: 14px;
  background: var(--bg);
  color: var(--text);
  transition: box-shadow 0.2s;
}

.btn:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}
</style>
