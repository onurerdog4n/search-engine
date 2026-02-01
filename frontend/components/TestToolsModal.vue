<template>
  <div v-if="show" class="modal-overlay" @click.self="close">
    <div class="modal-content">
      <div class="modal-header">
        <h2>🛠️ Test Araçları</h2>
        <button class="close-btn" @click="close">&times;</button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Sağlayıcı (Provider)</label>
          <select v-model="form.provider" class="input-field">
            <option value="provider-1">Provider 1 (JSON)</option>
            <option value="provider-2">Provider 2 (XML)</option>
          </select>
        </div>

        <div class="form-group">
          <label>İçerik Türü (Type)</label>
          <select v-model="form.type" class="input-field">
            <option value="video">Video</option>
            <option value="article">Makale (Article)</option>
          </select>
        </div>

        <div class="form-group">
          <label>İçerik ID (örn: json-v1, xml-a4)</label>
          <input v-model="form.id" type="text" class="input-field" placeholder="ID Girin" />
        </div>

        <div v-if="form.type === 'video'" class="grid-cols-2">
          <div class="form-group">
            <label>İzlenme (Views)</label>
            <input v-model.number="form.views" type="number" class="input-field" />
          </div>
          <div class="form-group">
            <label>Beğeni (Likes)</label>
            <input v-model.number="form.likes" type="number" class="input-field" />
          </div>
        </div>

        <div v-else class="grid-cols-2">
          <div class="form-group">
            <label>Okuma Süresi (Dakika)</label>
            <input v-model.number="form.reading_time" type="number" class="input-field" />
          </div>
          <div class="form-group">
            <label>Reaksiyonlar</label>
            <input v-model.number="form.reactions" type="number" class="input-field" />
          </div>
        </div>

        <div class="form-group">
          <label>Yayın Tarihi (Opsiyonel)</label>
          <input v-model="form.date" type="text" class="input-field" placeholder="2026-01-30T14:00:00Z" />
        </div>

        <div class="form-group">
          <label>Etiketler / Kategoriler (Virgülle ayırın)</label>
          <input v-model="form.tags" type="text" class="input-field" placeholder="örn: go, docker, cloud" />
        </div>

        <button class="btn btn-random" @click="randomizeData">
          🎲 Rastgele Veri Üret
        </button>

        <div class="action-buttons">
          <button class="btn btn-update" :disabled="loading" @click="updateData">
            {{ loading ? 'Güncelleniyor...' : 'Veriyi Güncelle' }}
          </button>
          <button class="btn btn-sync" :disabled="syncing" @click="triggerSync">
            {{ syncing ? 'Senkronize Ediliyor...' : 'Şimdi Senkronize Et' }}
          </button>
        </div>

        <div v-if="status" :class="['status-msg', status.type]">
          {{ status.message }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  show: Boolean
})
const emit = defineEmits(['close', 'synced'])

const form = ref({
  provider: 'provider-1',
  type: 'video',
  id: '',
  views: 1000,
  likes: 500,
  reading_time: 5,
  reactions: 50,
  date: '',
  tags: ''
})

const loading = ref(false)
const syncing = ref(false)
const status = ref(null)

const close = () => {
  status.value = null
  emit('close')
}

const randomizeData = () => {
  // Rastgele tür seç
  form.value.type = Math.random() > 0.3 ? 'video' : 'article'
  
  if (form.value.type === 'video') {
    form.value.views = Math.floor(Math.random() * 100000)
    form.value.likes = Math.floor(Math.random() * 5000)
    form.value.reading_time = 0
    form.value.reactions = 0
  } else {
    form.value.reading_time = Math.floor(Math.random() * 60) + 1
    form.value.reactions = Math.floor(Math.random() * 1000)
    form.value.views = 0
    form.value.likes = 0
  }
  
  // Son 4 ay içinde rastgele bir tarih
  const now = new Date()
  const randomDaysAgo = Math.floor(Math.random() * 120)
  const randomDate = new Date(now.getTime() - (randomDaysAgo * 24 * 60 * 60 * 1000))
  form.value.date = randomDate.toISOString()

  // Rastgele etiketler
  const availableTags = ['go', 'docker', 'kubernetes', 'cloud', 'backend', 'frontend', 'security', 'monitoring', 'tutorial']
  const shuffled = availableTags.sort(() => 0.5 - Math.random())
  form.value.tags = shuffled.slice(0, 3).join(', ')
}

const updateData = async () => {
  loading.value = true
  status.value = null
  try {
    const payload = {
      ...form.value,
      tags: form.value.tags ? form.value.tags.split(',').map(t => t.trim()).filter(t => t !== '') : []
    }
    const response = await fetch('http://localhost:8081/update-item', {
      method: 'POST',
      body: JSON.stringify(payload),
      headers: { 'Content-Type': 'application/json' }
    })
    
    if (response.ok) {
      status.value = { type: 'success', message: 'Veri başarıyla güncellendi!' }
    } else {
      status.value = { type: 'error', message: 'Güncelleme hatası!' }
    }
  } catch (e) {
    status.value = { type: 'error', message: 'Bağlantı hatası!' }
  } finally {
    loading.value = false
  }
}

const triggerSync = async () => {
  syncing.value = true
  status.value = null
  try {
    const response = await fetch('http://localhost:8080/api/v1/admin/sync', {
      method: 'POST'
    })
    
    if (response.ok) {
      status.value = { type: 'success', message: 'Senkronizasyon başlatıldı!' }
      setTimeout(() => {
        emit('synced')
      }, 2000)
    } else {
      status.value = { type: 'error', message: 'Senkronizasyon hatası!' }
    }
  } catch (e) {
    status.value = { type: 'error', message: 'Bağlantı hatası!' }
  } finally {
    syncing.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: #1c1c1e;
  width: 100%;
  max-width: 450px;
  border-radius: 12px;
  border: 1px solid #333;
  color: white;
  overflow: hidden;
}

.modal-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #333;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  color: #007aff;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 1.5rem;
  cursor: pointer;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.85rem;
  color: #888;
}

.input-field {
  width: 100%;
  padding: 0.75rem;
  background: #2c2c2e !important;
  border: 1px solid #444;
  border-radius: 8px;
  color: #ffffff !important;
  font-size: 0.9rem;
  outline: none;
  box-sizing: border-box;
}

.input-field:focus {
  border-color: #007aff;
  background: #3a3a3c !important;
}

/* Ensure select options are visible */
.input-field option {
  background: #2c2c2e;
  color: white;
}

.grid-cols-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn {
  padding: 0.75rem;
  border-radius: 8px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-update {
  background: #333;
  color: white;
}

.btn-update:hover {
  background: #444;
}

.btn-sync {
  background: #007aff;
  color: white;
}

.btn-sync:hover {
  background: #0063cc;
}

.btn-random {
  background: rgba(0, 122, 255, 0.1);
  color: #007aff;
  border: 1px solid rgba(0, 122, 255, 0.3);
  margin-bottom: 0.5rem;
  width: 100%;
}

.btn-random:hover {
  background: rgba(0, 122, 255, 0.2);
}

.status-msg {
  margin-top: 1rem;
  padding: 0.75rem;
  border-radius: 6px;
  font-size: 0.85rem;
  text-align: center;
}

.status-msg.success { background: rgba(52, 199, 89, 0.1); color: #34c759; }
.status-msg.error { background: rgba(255, 59, 48, 0.1); color: #ff3b30; }
</style>
