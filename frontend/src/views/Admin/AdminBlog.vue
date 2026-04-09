<template>
  <div class="admin-blog">
    <HelperBox title="AI Blog Generation" description="Generate blog content with AI. Requires an OpenAI API key." envVar="OPENAI_KEY" featureFlag="ai_blog_enabled" />
    <div class="page-header">
      <h1>Blog Posts</h1>
      <button @click="openCreate" class="btn btn-primary">+ New Post</button>
    </div>

    <div v-if="loading" class="loading">Loading posts...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <table v-else class="posts-table">
      <thead>
        <tr>
          <th>Title</th>
          <th>Status</th>
          <th>Date</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="posts.length === 0">
          <td colspan="4" class="empty-row">No posts yet.</td>
        </tr>
        <tr v-for="post in posts" :key="post.id">
          <td>{{ post.title }}</td>
          <td><span class="status-badge" :class="`status-${post.status}`">{{ post.status }}</span></td>
          <td>{{ formatDate(post.created_at) }}</td>
          <td class="actions">
            <button @click="openEdit(post)" class="btn btn-ghost btn-sm">Edit</button>
            <button @click="deletePost(post.id)" class="btn btn-danger btn-sm">Delete</button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Form Modal -->
    <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
      <div class="modal">
        <h2>{{ editingPost ? 'Edit Post' : 'New Post' }}</h2>

        <div class="form-group">
          <label>Title</label>
          <input v-model="form.title" type="text" class="input" placeholder="Post title" />
        </div>

        <div class="form-group">
          <label>Slug</label>
          <input v-model="form.slug" type="text" class="input" placeholder="post-url-slug" />
        </div>

        <div class="form-group">
          <label>Excerpt</label>
          <textarea v-model="form.excerpt" class="input textarea" rows="2" placeholder="Short description..."></textarea>
        </div>

        <div class="form-group">
          <label>Content</label>
          <textarea v-model="form.content" class="input textarea" rows="8" placeholder="Post content (HTML supported)..."></textarea>
        </div>

        <div class="form-group">
          <label>Status</label>
          <select v-model="form.status" class="input">
            <option value="draft">Draft</option>
            <option value="published">Published</option>
          </select>
        </div>

        <div class="form-group">
          <label>Cover Image URL</label>
          <div class="image-row">
            <input v-model="form.cover_image" type="text" class="input" placeholder="https://..." />
            <label class="btn btn-ghost upload-btn">
              Upload
              <input type="file" accept="image/*" @change="uploadCover" style="display:none" />
            </label>
          </div>
        </div>

        <div class="ai-section">
          <h3>Generate with AI</h3>
          <div class="ai-row">
            <input v-model="aiPrompt" type="text" class="input" placeholder="Describe the blog post..." />
            <button @click="generateWithAI" class="btn btn-ghost" :disabled="aiGenerating || !aiPrompt">
              {{ aiGenerating ? 'Generating...' : 'Generate' }}
            </button>
          </div>
          <p v-if="aiError" class="field-error">{{ aiError }}</p>
        </div>

        <p v-if="formError" class="field-error">{{ formError }}</p>

        <div class="modal-actions">
          <button @click="closeForm" class="btn btn-ghost">Cancel</button>
          <button @click="submitForm" class="btn btn-primary" :disabled="formSubmitting">
            {{ formSubmitting ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../../services/api'
import HelperBox from '../../components/admin/HelperBox.vue'

interface Post {
  id: string
  title: string
  slug: string
  excerpt: string
  content: string
  status: 'draft' | 'published'
  cover_image: string | null
  created_at: string
}

const posts = ref<Post[]>([])
const loading = ref(false)
const error = ref('')
const showForm = ref(false)
const editingPost = ref<Post | null>(null)
const formSubmitting = ref(false)
const formError = ref('')
const aiPrompt = ref('')
const aiGenerating = ref(false)
const aiError = ref('')

const form = ref({
  title: '',
  slug: '',
  excerpt: '',
  content: '',
  status: 'draft' as 'draft' | 'published',
  cover_image: '',
})

async function fetchPosts() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.get<{ data: Post[] }>('/api/v1/admin/blog')
    posts.value = data.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load posts'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingPost.value = null
  form.value = { title: '', slug: '', excerpt: '', content: '', status: 'draft', cover_image: '' }
  formError.value = ''
  showForm.value = true
}

function openEdit(post: Post) {
  editingPost.value = post
  form.value = {
    title: post.title,
    slug: post.slug,
    excerpt: post.excerpt,
    content: post.content,
    status: post.status,
    cover_image: post.cover_image || '',
  }
  formError.value = ''
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingPost.value = null
}

async function submitForm() {
  formSubmitting.value = true
  formError.value = ''
  try {
    const payload = { ...form.value, cover_image: form.value.cover_image || null }
    if (editingPost.value) {
      await api.put(`/api/v1/admin/blog/${editingPost.value.id}`, payload)
    } else {
      await api.post('/api/v1/admin/blog', payload)
    }
    await fetchPosts()
    closeForm()
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to save post'
  } finally {
    formSubmitting.value = false
  }
}

async function deletePost(id: string) {
  if (!confirm('Delete this post?')) return
  try {
    await api.delete(`/api/v1/admin/blog/${id}`)
    await fetchPosts()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete post'
  }
}

async function generateWithAI() {
  aiGenerating.value = true
  aiError.value = ''
  try {
    const data = await api.post<{ title: string; content: string; excerpt: string; slug: string }>(
      '/api/v1/admin/blog/ai-generate',
      { prompt: aiPrompt.value }
    )
    if (data.title) form.value.title = data.title
    if (data.content) form.value.content = data.content
    if (data.excerpt) form.value.excerpt = data.excerpt
    if (data.slug) form.value.slug = data.slug
    aiPrompt.value = ''
  } catch (e: unknown) {
    aiError.value = e instanceof Error ? e.message : 'AI generation failed'
  } finally {
    aiGenerating.value = false
  }
}

async function uploadCover(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  try {
    const data = await fetch(`${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/api/v1/admin/upload`, {
      method: 'POST',
      credentials: 'include',
      body: formData,
    }).then(r => r.json()) as { url: string }
    form.value.cover_image = data.url
  } catch {
    formError.value = 'Failed to upload image'
  }
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(fetchPosts)
</script>

<style scoped>
.admin-blog {
  padding: 32px;
  text-align: left;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
}

.loading, .empty-row {
  text-align: center;
  color: var(--text);
  padding: 32px;
}

.error {
  color: #ef4444;
  padding: 12px;
  border: 1px solid #ef4444;
  border-radius: 6px;
  margin-bottom: 16px;
}

.posts-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.posts-table th,
.posts-table td {
  text-align: left;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.posts-table th {
  color: var(--text);
  font-weight: 500;
  background: var(--code-bg);
}

.posts-table tr:hover td {
  background: var(--social-bg);
}

.status-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
  text-transform: capitalize;
}

.status-draft { background: rgba(107, 114, 128, 0.15); color: #6b7280; }
.status-published { background: rgba(34, 197, 94, 0.15); color: #16a34a; }

.actions {
  display: flex;
  gap: 8px;
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
  font-family: inherit;
}

.btn-sm {
  padding: 4px 10px;
  font-size: 13px;
}

.btn-primary {
  background: var(--accent);
  color: white;
  border-color: var(--accent);
}

.btn-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border-color: rgba(239, 68, 68, 0.3);
}

.btn-ghost {
  background: transparent;
}

.btn:hover:not(:disabled) {
  box-shadow: var(--shadow);
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 32px 16px;
  z-index: 100;
  overflow-y: auto;
}

.modal {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 32px;
  width: 100%;
  max-width: 640px;
}

.modal h2 {
  margin: 0 0 24px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--text);
  margin-bottom: 6px;
}

.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--border);
  border-radius: 6px;
  background: var(--bg);
  color: var(--text-h);
  font-size: 14px;
  box-sizing: border-box;
  font-family: inherit;
}

.textarea {
  resize: vertical;
  min-height: 80px;
}

.input:focus {
  outline: 2px solid var(--accent);
  outline-offset: -1px;
}

.image-row {
  display: flex;
  gap: 8px;
}

.upload-btn {
  cursor: pointer;
  white-space: nowrap;
}

.ai-section {
  border-top: 1px solid var(--border);
  padding-top: 16px;
  margin: 16px 0;
}

.ai-section h3 {
  margin: 0 0 12px;
  font-size: 15px;
  color: var(--text-h);
}

.ai-row {
  display: flex;
  gap: 8px;
}

.field-error {
  color: #ef4444;
  font-size: 13px;
  margin: 8px 0 0;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  border-top: 1px solid var(--border);
  padding-top: 20px;
}
</style>
