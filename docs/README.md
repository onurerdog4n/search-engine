# 📚 Dokümantasyon Sitesi

Bu proje için interaktif dokümantasyon sitesi. **Nuxt Content** ile oluşturulmuştur.

## 🚀 Çalıştırma

```bash
# Dependencies yükle
npm install

# Development server başlat
npm run dev

# Production build
npm run build
npm run preview
```

## 📖 Dokümantasyon İçeriği

### Genel Bakış
- **Ana Sayfa**: Proje genel bakış
- **Mimari Tasarım**: Clean Architecture ve katman yapısı
- **Özellikler**: Temel özellikler ve yetenekler

### Teknik Detaylar
- **Veri Akışı**: End-to-end veri akışı diyagramları
- **Performans**: Optimizasyon teknikleri ve metrikler
- **API Referansı**: REST API endpoint'leri ve kullanımı

### Kalite & Güvenlik
- **Test Coverage**: Unit, integration ve e2e testler (%75+ coverage)
- **Monitoring & Logging**: Prometheus metrics, Zap logging, pprof profiling
- **Security**: OWASP Top 10, input validation, rate limiting

### Başlangıç
- **Kurulum**: Docker Compose ile kurulum adımları

## 📁 Dosya Yapısı

```
docs/
├── app/
│   └── layouts/
│       └── default.vue          # Ana layout (sidebar + content)
├── content/
│   └── tr/                      # Türkçe içerik
│       ├── index.md             # Ana sayfa
│       ├── architecture.md      # Mimari
│       ├── features.md          # Özellikler
│       ├── data-flow.md         # Veri akışı
│       ├── performance.md       # Performans
│       ├── api.md               # API referansı
│       ├── testing.md           # Test coverage ✨ YENİ
│       ├── monitoring.md        # Monitoring ✨ YENİ
│       ├── security.md          # Security ✨ YENİ
│       └── installation.md      # Kurulum
├── TURKISH_DOCUMENTATION.md     # Tek dosya dokümantasyon
└── README.md                    # Bu dosya
```

## 🎨 Özellikler

- ✅ **Responsive Design**: Mobil ve desktop uyumlu
- ✅ **Dark Mode Ready**: Kolay dark mode entegrasyonu
- ✅ **Syntax Highlighting**: Code block'lar için syntax highlighting
- ✅ **Search**: Dokümantasyon içinde arama (gelecek)
- ✅ **Auto Navigation**: Otomatik sidebar navigation

## 🔧 Teknolojiler

- **Nuxt 3**: Vue.js framework
- **Nuxt Content**: Markdown-based content management
- **Vue 3**: Progressive JavaScript framework
- **TypeScript**: Type-safe development

## 📝 Yeni İçerik Ekleme

1. `content/tr/` klasörüne yeni `.md` dosyası ekle
2. Frontmatter ekle:
```yaml
---
title: "Başlık"
description: "Açıklama"
---
```
3. Markdown içeriğini yaz
4. `app/layouts/default.vue` dosyasına menü linki ekle

## 🌐 Canlı Önizleme

Development server: http://localhost:3000

## 📚 Kaynaklar

- [Nuxt Content Documentation](https://content.nuxtjs.org/)
- [Nuxt 3 Documentation](https://nuxt.com/)
- [Markdown Guide](https://www.markdownguide.org/)
