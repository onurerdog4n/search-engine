---
title: API Referansı
description: RESTful API endpoint dokümantasyonu, request/response örnekleri ve kullanım senaryoları
navigation: true
---

# 📡 API Referansı

RESTful API endpoint'lerinin detaylı dokümantasyonu.

## Base URL

```
Development:  http://localhost:8080/api/v1
Production:   https://api.yourdomain.com/api/v1
```

## Endpoints

### 1. 🔍 Search - Arama Endpoint'i

İçeriklerde arama yapar, filtreler ve sıralar.

#### Request

```http
GET /api/v1/search
```

#### Query Parameters

| Parametre | Tip | Zorunlu | Default | Açıklama |
|-----------|-----|---------|---------|----------|
| `query` | string | ❌ | `""` | Arama terimi (boş ise tüm sonuçlar) |
| `type` | string | ❌ | `""` | `video` veya `article` |
| `sort` | string | ❌ | `popularity` | `popularity`, `relevance` veya `date` |
| `page` | integer | ❌ | `1` | Sayfa numarası (min: 1, max: 1000) |
| `page_size` | integer | ❌ | `20` | Sayfa boyutu (min: 1, max: 100) |

#### Response

**Success (200 OK):**

```json
{
  "items": [
    {
      "id": 1,
      "provider_id": 1,
      "title": "Go Programming Tutorial for Beginners",
      "description": "Learn Go from scratch with practical examples",
      "content_type": "video",
      "published_at": "2024-01-15T10:00:00Z",
      "stats": {
        "views": 150000,
        "likes": 5000,
        "reading_time": 0,
        "reactions": 0
      },
      "score": {
        "base_score": 200.0,
        "type_weight": 1.5,
        "recency_score": 5.0,
        "engagement_score": 3.33,
        "final_score": 308.33,
        "calculated_at": "2024-01-20T14:30:00Z"
      },
      "tags": [
        {"id": 1, "name": "golang"},
        {"id": 2, "name": "tutorial"},
        {"id": 3, "name": "beginner"}
      ],
      "relevance_score": 0.95,
      "created_at": "2024-01-15T10:05:00Z",
      "updated_at": "2024-01-20T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 150,
    "total_pages": 8
  }
}
```

#### Kullanım Örnekleri

::code-group

```bash [cURL]
# Basit arama
curl "http://localhost:8080/api/v1/search?query=golang"

# Video içeriklerde arama
curl "http://localhost:8080/api/v1/search?query=tutorial&type=video"

# Popülerliğe göre sıralama
curl "http://localhost:8080/api/v1/search?query=go&sort=popularity&page=1&page_size=10"

# Alakalılığa göre sıralama
curl "http://localhost:8080/api/v1/search?query=go&sort=relevance"
```

```javascript [JavaScript/Fetch]
async function searchContents(query, filters = {}) {
  const params = new URLSearchParams({
    query: query,
    ...filters
  });
  
  const response = await fetch(
    `http://localhost:8080/api/v1/search?${params}`
  );
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const data = await response.json();
  return data;
}

// Kullanım
const results = await searchContents('golang', {
  type: 'video',
  sort: 'popularity',
  page: 1,
  page_size: 20
});

console.log(`Found ${results.pagination.total_items} items`);
results.items.forEach(item => {
  console.log(`- ${item.title} (Score: ${item.score.final_score})`);
});
```

```go [Go]
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

type SearchResult struct {
    Items      []Content   `json:"items"`
    Pagination Pagination `json:"pagination"`
}

func search(query, contentType, sortBy string, page, pageSize int) (*SearchResult, error) {
    baseURL := "http://localhost:8080/api/v1/search"
    
    params := url.Values{}
    params.Add("query", query)
    if contentType != "" {
        params.Add("type", contentType)
    }
    params.Add("sort", sortBy)
    params.Add("page", fmt.Sprintf("%d", page))
    params.Add("page_size", fmt.Sprintf("%d", pageSize))
    
    resp, err := http.Get(fmt.Sprintf("%s?%s", baseURL, params.Encode()))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result SearchResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}

func main() {
    result, err := search("golang", "video", "popularity", 1, 20)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d items\n", result.Pagination.TotalItems)
    for _, item := range result.Items {
        fmt.Printf("- %s (Score: %.2f)\n", 
            item.Title, 
            item.Score.FinalScore)
    }
}
```

```python [Python]
import requests
from typing import Optional, Dict, List

class SearchClient:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
    
    def search(
        self,
        query: str = "",
        content_type: Optional[str] = None,
        sort_by: str = "popularity",
        page: int = 1,
        page_size: int = 20
    ) -> Dict:
        params = {
            "query": query,
            "sort": sort_by,
            "page": page,
            "page_size": page_size
        }
        
        if content_type:
            params["type"] = content_type
        
        response = requests.get(
            f"{self.base_url}/search",
            params=params
        )
        response.raise_for_status()
        
        return response.json()

# Kullanım
client = SearchClient()

# Video arama
results = client.search(
    query="golang",
    content_type="video",
    sort_by="popularity"
)

print(f"Found {results['pagination']['total_items']} items")
for item in results['items']:
    print(f"- {item['title']} (Score: {item['score']['final_score']:.2f})")
```

::

#### Kullanım Senaryoları

**1. Genel Arama (Tüm İçerikler)**
```bash
GET /api/v1/search?query=programming
```

**2. Sadece Videolar**
```bash
GET /api/v1/search?query=tutorial&type=video
```

**3. Alakalılığa Göre Sıralama**
```bash
GET /api/v1/search?query=golang%20tutorial&sort=relevance
```

**4. Pagination**
```bash
# 2. sayfa, her sayfada 50 sonuç
GET /api/v1/search?query=go&page=2&page_size=50
```

**5. Tüm İçerikleri Listeleme**
```bash
# query boş = tüm sonuçlar
GET /api/v1/search?sort=date&page_size=100
```

### 2. 🔄 Admin Sync - Manuel Senkronizasyon

Provider'lardan manuel veri senkronizasyonu başlatır.

#### Request

```http
POST /api/v1/admin/sync
Content-Type: application/json
```

#### Response

**Success (200 OK):**

```json
{
  "message": "Senkronizasyon başlatıldı",
  "started_at": "2024-01-20T14:30:00Z"
}
```

#### Kullanım Örnekleri

::code-group

```bash [cURL]
curl -X POST http://localhost:8080/api/v1/admin/sync
```

```javascript [JavaScript]
async function triggerSync() {
  const response = await fetch('http://localhost:8080/api/v1/admin/sync', {
    method: 'POST'
  });
  
  const data = await response.json();
  console.log(data.message);
}

await triggerSync();
```

```go [Go]
resp, err := http.Post(
    "http://localhost:8080/api/v1/admin/sync",
    "application/json",
    nil,
)
```

::

::alert{type="warning"}
**Production:** Bu endpoint authentication gerektirir. JWT token veya API key ile korunmalıdır.
::

**Otomatik Senkronizasyon:**

Sistem varsayılan olarak **saatte bir** otomatik senkronizasyon yapar. Manuel sync sadece acil durumlar için.

### 3. ❤️ Health Check

Servis sağlığını kontrol eder.

#### Request

```http
GET /api/v1/health
```

#### Response

**Healthy (200 OK):**

```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T14:30:00Z",
  "version": "1.0.0",
  "uptime_seconds": 3600
}
```

**Unhealthy (503 Service Unavailable):**

```json
{
  "status": "unhealthy",
  "timestamp": "2024-01-20T14:30:00Z",
  "errors": [
    "Database connection failed",
    "Redis unavailable"
  ]
}
```

#### Kullanım

```bash
# Basit health check
curl http://localhost:8080/api/v1/health

# Kubernetes liveness probe
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

## 🔐 Güvenlik

### Rate Limiting

API, **IP bazlı rate limiting** kullanır:

- **Limit:** 60 istek/dakika
- **Status Code:** 429 Too Many Requests
- **Retry Header:** `Retry-After: 60`

**Response Headers:**

```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1706019000
```

**Rate Limit Aşıldığında:**

```json
HTTP/1.1 429 Too Many Requests
Retry-After: 60
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0

{
  "error": "Rate limit exceeded. Please try again in 60 seconds."
}
```

### CORS

API CORS yapılandırması:

**Allowed Origins:**
```
Development: http://localhost:3000, http://localhost:8080
Production:  https://yourdomain.com
```

**Allowed Methods:**
```
GET, POST, OPTIONS
```

**Allowed Headers:**
```
Content-Type, Authorization, X-Request-ID
```

**Preflight Request:**
```http
OPTIONS /api/v1/search HTTP/1.1
Origin: http://localhost:3000
Access-Control-Request-Method: GET

→ 200 OK
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Max-Age: 3600
```

## ❌ Error Responses

### 400 Bad Request

Geçersiz parametreler:

```json
{
  "error": "Invalid parameter",
  "details": {
    "field": "sort",
    "message": "sort must be one of: popularity, relevance, date"
  }
}
```

**Örnekler:**
- Geçersiz `sort` değeri
- `page` < 1 veya `page` > 1000
- `page_size` > 100
- Geçersiz `type` (video, article dışında)

### 404 Not Found

Endpoint or resource bulunamadı:

```json
{
  "error": "Not found",
  "path": "/api/v1/invalid"
}
```

### 429 Too Many Requests

Rate limit aşıldı:

```json
{
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

### 500 Internal Server Error

Sunucu hatası:

```json
{
  "error": "Internal server error",
  "request_id": "abc-123-def"
}
```

::alert{type="info"}
Tüm error response'lar `request_id` içerir. Bu ID ile log'lardan detay bulunabilir.
::

## 📋 Data Models

### Content

```typescript
interface Content {
  id: number;
  provider_id: number;
  title: string;
  description: string;
  content_type: "video" | "article";
  published_at: string;  // ISO 8601
  stats?: ContentStats;
  score?: ContentScore;
  tags?: Tag[];
  relevance_score?: number;  // Sadece arama sonuçlarında
  created_at: string;
  updated_at: string;
}
```

### ContentStats

```typescript
interface ContentStats {
  views: number;
  likes: number;
  reading_time: number;  // Dakika
  reactions: number;
}
```

### ContentScore

```typescript
interface ContentScore {
  base_score: number;
  type_weight: number;
  recency_score: number;
  engagement_score: number;
  final_score: number;
  calculated_at: string;  // ISO 8601
}
```

### Tag

```typescript
interface Tag {
  id: number;
  name: string;
}
```

### Pagination

```typescript
interface Pagination {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}
```

### SearchResult

```typescript
interface SearchResult {
  items: Content[];
  pagination: Pagination;
}
```

## 🧪 Testing Endpoints

### Postman Collection

```json
{
  "info": {
    "name": "Search Engine API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Search - All",
      "request": {
        "method": "GET",
        "url": {
          "raw": "{{baseUrl}}/search?query=golang",
          "host": ["{{baseUrl}}"],
          "path": ["search"],
          "query": [
            {"key": "query", "value": "golang"}
          ]
        }
      }
    },
    {
      "name": "Search - Videos Only",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/search?query=tutorial&type=video&sort=popularity"
      }
    },
    {
      "name": "Admin Sync",
      "request": {
        "method": "POST",
        "url": "{{baseUrl}}/admin/sync"
      }
    },
    {
      "name": "Health Check",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/health"
      }
    }
  ],
  "variable": [
    {
      "key": "baseUrl",
      "value": "http://localhost:8080/api/v1"
    }
  ]
}
```

### HTTPie Examples

```bash
# Search
http GET localhost:8080/api/v1/search query==golang type==video

# Sync
http POST localhost:8080/api/v1/admin/sync

# Health
http GET localhost:8080/api/v1/health
```

## 📊 Response Times

| Endpoint | Cache Hit | Cache Miss | Target |
|----------|-----------|------------|--------|
| `/search` | 15-25ms | 80-150ms | <500ms |
| `/admin/sync` | N/A | 2-5s | <30s |
| `/health` | 1-5ms | N/A | <10ms |

## 🚀 Best Practices

::list{type="success"}
- **Pagination:** Always use pagination for large result sets
- **Caching:** Use same parameters to benefit from cache
- **Rate Limiting:** Implement client-side rate limiting
- **Error Handling:** Always check status codes and handle errors
- **Request ID:** Log request_id for debugging
::

## 📚 OpenAPI Specification

Full OpenAPI 3.0 specification available at:

```
GET /api/v1/openapi.json
```

Import to Swagger UI, Postman, or Insomnia for interactive API testing.
