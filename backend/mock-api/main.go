package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

type JSONMetrics struct {
	Views       *int64  `json:"views,omitempty"`
	Likes       *int32  `json:"likes,omitempty"`
	Duration    string  `json:"duration,omitempty"`
	ReadingTime *int32  `json:"reading_time,omitempty"`
	Reactions   *int32  `json:"reactions,omitempty"`
}

type JSONContent struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Metrics     JSONMetrics `json:"metrics"`
	PublishedAt string      `json:"published_at"`
	Tags        []string    `json:"tags"`
}

type JSONResponse struct {
	Contents   []JSONContent `json:"contents"`
	Pagination struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"pagination"`
}

type XMLRoot struct {
	XMLName xml.Name `xml:"feed"`
	Items   struct {
		Items []XMLContent `xml:"item"`
	} `xml:"items"`
	Meta struct {
		TotalCount   int `xml:"total_count"`
		CurrentPage  int `xml:"current_page"`
		ItemsPerPage int `xml:"items_per_page"`
	} `xml:"meta"`
}

type XMLContent struct {
	ID              string `xml:"id"`
	Headline        string `xml:"headline"`
	Type            string `xml:"type"`
	Stats           struct {
		Views       *int64  `xml:"views,omitempty"`
		Likes       *int32  `xml:"likes,omitempty"`
		Duration    string  `xml:"duration,omitempty"`
		ReadingTime *int32  `xml:"reading_time,omitempty"`
		Reactions   *int32  `xml:"reactions,omitempty"`
		Comments    *int32  `xml:"comments,omitempty"`
	} `xml:"stats"`
	PublicationDate string `xml:"publication_date"`
	Categories      struct {
		Categories []string `xml:"category"`
	} `xml:"categories"`
}

type UpdateItemRequest struct {
	Provider    string   `json:"provider"` // "provider-1" or "provider-2"
	ID          string   `json:"id"`
	Views       int64    `json:"views"`
	Likes       int32    `json:"likes"`
	ReadingTime int32    `json:"reading_time"`
	Reactions   int32    `json:"reactions"`
	Date        string   `json:"date"` // YYYY-MM-DD or RFC3339
	Tags        []string `json:"tags"`
}

func main() {
	http.HandleFunc("/provider-1", enableCORS(handleJSON))
	http.HandleFunc("/provider-2", enableCORS(handleXML))
	http.HandleFunc("/update-item", enableCORS(handleUpdateItem))

	port := ":8081"
	fmt.Printf("Mock API server starting on %s...\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Provider == "provider-1" {
		data, _ := os.ReadFile("/app/mocks/provider1.json")
		var resp JSONResponse
		json.Unmarshal(data, &resp)

		found := false
		for i, item := range resp.Contents {
			if item.ID == req.ID {
				if item.Type == "video" {
					resp.Contents[i].Metrics.Views = &req.Views
					resp.Contents[i].Metrics.Likes = &req.Likes
					resp.Contents[i].Metrics.ReadingTime = nil
					resp.Contents[i].Metrics.Reactions = nil
				} else {
					resp.Contents[i].Metrics.ReadingTime = &req.ReadingTime
					resp.Contents[i].Metrics.Reactions = &req.Reactions
					resp.Contents[i].Metrics.Views = nil
					resp.Contents[i].Metrics.Likes = nil
					resp.Contents[i].Metrics.Duration = ""
				}

				if req.Date != "" {
					resp.Contents[i].PublishedAt = req.Date
				}
				if len(req.Tags) > 0 {
					resp.Contents[i].Tags = req.Tags
				}
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		newData, _ := json.MarshalIndent(resp, "", "  ")
		os.WriteFile("/app/mocks/provider1.json", newData, 0644)

	} else if req.Provider == "provider-2" {
		data, _ := os.ReadFile("/app/mocks/provider2.xml")
		var resp XMLRoot
		xml.Unmarshal(data, &resp)

		found := false
		for i, item := range resp.Items.Items {
			if item.ID == req.ID {
				if item.Type == "video" {
					resp.Items.Items[i].Stats.Views = &req.Views
					resp.Items.Items[i].Stats.Likes = &req.Likes
					resp.Items.Items[i].Stats.ReadingTime = nil
					resp.Items.Items[i].Stats.Reactions = nil
					resp.Items.Items[i].Stats.Comments = nil
				} else {
					resp.Items.Items[i].Stats.ReadingTime = &req.ReadingTime
					resp.Items.Items[i].Stats.Reactions = &req.Reactions
					resp.Items.Items[i].Stats.Views = nil
					resp.Items.Items[i].Stats.Likes = nil
					resp.Items.Items[i].Stats.Duration = ""
				}

				if req.Date != "" {
					resp.Items.Items[i].PublicationDate = req.Date
				}
				if len(req.Tags) > 0 {
					resp.Items.Items[i].Categories.Categories = req.Tags
				}
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		newData, _ := xml.MarshalIndent(resp, "", "  ")
		os.WriteFile("/app/mocks/provider2.xml", []byte(xml.Header+string(newData)), 0644)
	} else {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("/app/mocks/provider1.json")
	if err != nil {
		http.Error(w, "File not found", 500)
		return
	}

	var fullResponse JSONResponse
	json.Unmarshal(data, &fullResponse)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	totalItems := len(fullResponse.Contents)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > totalItems {
		fullResponse.Contents = []JSONContent{}
	} else {
		if end > totalItems {
			end = totalItems
		}
		fullResponse.Contents = fullResponse.Contents[start:end]
	}

	fullResponse.Pagination.Page = page
	fullResponse.Pagination.PerPage = pageSize
	fullResponse.Pagination.Total = totalItems

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullResponse)
}

func handleXML(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("/app/mocks/provider2.xml")
	if err != nil {
		http.Error(w, "File not found", 500)
		return
	}

	var fullResponse XMLRoot
	xml.Unmarshal(data, &fullResponse)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 10

	totalItems := len(fullResponse.Items.Items)
	start := (page - 1) * pageSize
	end := start + pageSize

	var slicedItems []XMLContent
	if start < totalItems {
		if end > totalItems {
			end = totalItems
		}
		slicedItems = fullResponse.Items.Items[start:end]
	}
	fullResponse.Items.Items = slicedItems
	fullResponse.Meta.CurrentPage = page
	fullResponse.Meta.ItemsPerPage = pageSize
	fullResponse.Meta.TotalCount = totalItems

	w.Header().Set("Content-Type", "application/xml")
	xml.NewEncoder(w).Encode(fullResponse)
}
