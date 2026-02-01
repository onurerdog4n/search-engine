<template>
  <div class="doc-layout">
    <aside class="sidebar">
      <div class="logo">
        <NuxtLink to="/">
          <h2>🔍 Arama Motoru</h2>
        </NuxtLink>
      </div>
      
      <nav>
        <div class="nav-section">
          <h3 class="nav-title">Genel Bakış</h3>
          <NuxtLink to="/" class="nav-link">
            <span class="icon">🏠</span>
            Ana Sayfa
          </NuxtLink>
          <NuxtLink to="/architecture" class="nav-link">
            <span class="icon">🏗️</span>
            Mimari Tasarım
          </NuxtLink>
          <NuxtLink to="/features" class="nav-link">
            <span class="icon">⚡</span>
            Özellikler
          </NuxtLink>
        </div>

        <div class="nav-section">
          <h3 class="nav-title">Teknik Detaylar</h3>
          <NuxtLink to="/data-flow" class="nav-link">
            <span class="icon">🔄</span>
            Veri Akışı
          </NuxtLink>
          <NuxtLink to="/performance" class="nav-link">
            <span class="icon">🚀</span>
            Performans
          </NuxtLink>
          <NuxtLink to="/api" class="nav-link">
            <span class="icon">📡</span>
            API Referansı
          </NuxtLink>
        </div>

        <div class="nav-section">
          <h3 class="nav-title">Kalite & Güvenlik</h3>
          <NuxtLink to="/testing" class="nav-link">
            <span class="icon">🧪</span>
            Test Coverage
          </NuxtLink>
          <NuxtLink to="/monitoring" class="nav-link">
            <span class="icon">📊</span>
            Monitoring & Logging
          </NuxtLink>
          <NuxtLink to="/security" class="nav-link">
            <span class="icon">🔒</span>
            Security
          </NuxtLink>
        </div>

        <div class="nav-section">
          <h3 class="nav-title">Başlangıç</h3>
          <NuxtLink to="/installation" class="nav-link">
            <span class="icon">📦</span>
            Kurulum
          </NuxtLink>
        </div>
      </nav>

      <div class="sidebar-footer">
        <a href="https://github.com/onurerdog4n" target="_blank" class="github-link">
          <span class="icon">💻</span>
          GitHub
        </a>
      </div>
    </aside>
    
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.doc-layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 280px;
  background: #f8f9fa;
  border-right: 1px solid #e1e4e8;
  padding: 2rem 0;
  position: fixed;
  height: 100vh;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.logo {
  padding: 0 1.5rem 2rem;
  border-bottom: 1px solid #e1e4e8;
  margin-bottom: 1.5rem;
}

.logo h2 {
  margin: 0;
  font-size: 1.25rem;
  color: #1a1a1a;
  font-weight: 600;
}

.logo a {
  text-decoration: none;
  color: inherit;
}

nav {
  flex: 1;
  padding: 0 1rem;
}

.nav-section {
  margin-bottom: 2rem;
}

.nav-title {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #6a737d;
  font-weight: 600;
  padding: 0 0.5rem;
  margin-bottom: 0.5rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  color: #24292e;
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s;
  font-size: 0.925rem;
  margin-bottom: 0.25rem;
}

.nav-link:hover {
  background: #e1e4e8;
  color: #0969da;
}

.nav-link.router-link-active {
  background: #0969da;
  color: white;
  font-weight: 500;
}

.icon {
  font-size: 1.1rem;
  opacity: 0.8;
}

.sidebar-footer {
  padding: 1.5rem;
  border-top: 1px solid #e1e4e8;
}

.github-link {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  color: #24292e;
  text-decoration: none;
  border-radius: 6px;
  background: #f6f8fa;
  transition: all 0.2s;
  font-size: 0.925rem;
  font-weight: 500;
}

.github-link:hover {
  background: #0969da;
  color: white;
}

.main-content {
  margin-left: 280px;
  flex: 1;
  padding: 3rem;
  max-width: 1200px;
}

/* Responsive */
@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    position: relative;
    height: auto;
  }
  
  .main-content {
    margin-left: 0;
    padding: 1.5rem;
  }
}
</style>
