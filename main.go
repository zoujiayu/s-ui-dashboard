package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sort"
	"time"
)

const (
	BaseURL = "http://127.0.0.1:2095/app/apiv2"
	Token   = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
)

// ---------- Data Structures ----------
type StatusResp struct {
	Obj struct {
		SBD struct {
			Running bool `json:"running"`
			Stats   struct {
				Alloc        int64 `json:"Alloc"`
				NumGoroutine int   `json:"NumGoroutine"`
				Uptime       int64 `json:"Uptime"`
			} `json:"stats"`
		} `json:"sbd"`
		Mem struct {
			Current int64 `json:"current"`
			Total   int64 `json:"total"`
		} `json:"mem"`
		Net struct {
			Recv int64 `json:"recv"`
			Sent int64 `json:"sent"`
		} `json:"net"`
		Uptime int64 `json:"uptime"`
	} `json:"obj"`
}

type Client struct {
	Name   string `json:"name"`
	Down   int64  `json:"down"`
	Up     int64  `json:"up"`
	Volume int64  `json:"volume"` // 0 = unlimited
	Expiry int64  `json:"expiry"`
}

type ClientsResp struct {
	Obj struct {
		Clients []Client `json:"clients"`
	} `json:"obj"`
}

type RealTimeTraffic struct {
	ID        int64  `json:"id"`
	DateTime  int64  `json:"dateTime"`
	Resource  string `json:"resource"`
	Tag       string `json:"tag"`
	Direction bool   `json:"direction"` // true = upload
	Traffic   int64  `json:"traffic"`   // bytes
}

type TrafficResp struct {
	Obj []RealTimeTraffic `json:"obj"`
}

// Aggregated data for frontend chart
type TrafficPoint struct {
	Time int64   `json:"time"`
	Up   float64 `json:"up"`   // MB
	Down float64 `json:"down"` // MB
}

type PageData struct {
	User      string
	Remaining string
	Runtime   string
	MemUsedMB int
	MemPct    float64
	NetRxGB   float64
	NetTxGB   float64
	UpGB      float64
	DownGB    float64
	RemainStr string
	TotalStr  string
	Expiry    string
}

// ---------- HTTP API Request ----------
func apiGet(url string, v any) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Token", Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return json.Unmarshal(body, v)
}

// ---------- Aggregate Real-Time Traffic ----------
func ParseTraffic(resp TrafficResp) []TrafficPoint {
	pointsMap := map[int64]*TrafficPoint{}
	for _, t := range resp.Obj {
		p, ok := pointsMap[t.DateTime]
		if !ok {
			p = &TrafficPoint{Time: t.DateTime}
			pointsMap[t.DateTime] = p
		}
		if t.Direction {
			p.Up += float64(t.Traffic) / 1024 / 1024
		} else {
			p.Down += float64(t.Traffic) / 1024 / 1024
		}
	}
	points := []TrafficPoint{}
	for _, p := range pointsMap {
		points = append(points, *p)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Time < points[j].Time })
	return points
}

func formatRuntime(seconds int64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	return fmt.Sprintf("%d days %d hours", days, hours)
}

// ---------- Main ----------
func main() {
	tpl := template.Must(template.ParseFiles("template/template.html"))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		userName := r.URL.Query().Get("user")
		if userName == "" {
			http.NotFound(w, r)
			return
		}

		// System status
		var status StatusResp
		if err := apiGet(BaseURL+"/status?r=sys,sbd", &status); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Clients list
		var clients ClientsResp
		if err := apiGet(BaseURL+"/clients", &clients); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var user Client
		found := false
		for _, c := range clients.Obj.Clients {
			if c.Name == userName {
				user = c
				found = true
				break
			}
		}
		if !found {
			http.NotFound(w, r)
			return
		}

		remainStr := "Unlimited"
		totalStr := "Unlimited"
		expiry := "Unlimited"
		if user.Volume > 0 {
			totalGB := float64(user.Volume) / 1024 / 1024 / 1024
			remainGB := float64(user.Volume-user.Down-user.Up) / 1024 / 1024 / 1024
			remainStr = fmt.Sprintf("%.2f GB", remainGB)
			totalStr = fmt.Sprintf("%.2f GB", totalGB)
			expiry = time.Unix(user.Expiry, 0).Format("2006-01-02 15:04:05")
		}

		data := PageData{
			User:      user.Name,
			Remaining: remainStr,
			MemUsedMB: int(status.Obj.Mem.Current / 1024 / 1024),
			MemPct:    float64(status.Obj.Mem.Current) / float64(status.Obj.Mem.Total) * 100,
			NetRxGB:   float64(status.Obj.Net.Recv) / 1024 / 1024 / 1024,
			NetTxGB:   float64(status.Obj.Net.Sent) / 1024 / 1024 / 1024,
			Runtime:   formatRuntime(status.Obj.SBD.Stats.Uptime),
			UpGB:      float64(user.Up) / 1024 / 1024 / 1024,
			DownGB:    float64(user.Down) / 1024 / 1024 / 1024,
			TotalStr:  totalStr,
			Expiry:    expiry,
		}

		tpl.Execute(w, data)
	})

	// ---------- Async Traffic API ----------
	http.HandleFunc("/api/traffic", func(w http.ResponseWriter, r *http.Request) {
		userName := r.URL.Query().Get("user")
		if userName == "" {
			http.NotFound(w, r)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		if limitStr == "" {
			limitStr = "6"
		}

		var limit int
		switch limitStr {
		case "1":
			limit = 1
		case "6":
			limit = 6
		case "12":
			limit = 12
		case "1d":
			limit = 24
		case "7d":
			limit = 24 * 7
		case "15d":
			limit = 24 * 15
		case "30d":
			limit = 24 * 30
		default:
			limit = 6
		}

		url := fmt.Sprintf("%s/stats?resource=user&tag=%s&limit=%d", BaseURL, userName, limit)
		var resp TrafficResp
		if err := apiGet(url, &resp); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		points := ParseTraffic(resp)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(points)
	})

	log.Println("Listening on :2097")
	srv := &http.Server{
		Addr:         ":2097",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
