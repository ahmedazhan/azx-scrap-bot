package db

import (
	"strings"
	"time"
)

type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64" json:"username"`
	PasswordHash string    `gorm:"type:text" json:"-"`
	Theme        string    `gorm:"size:16;default:'dark'" json:"theme"`
	FilterMode   string    `gorm:"size:16;default:'sheet'" json:"filter_mode"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Source struct {
	ID               uint64     `gorm:"primaryKey" json:"id"`
	Key              string     `gorm:"uniqueIndex;size:64" json:"key"`
	DisplayName      string     `gorm:"size:128" json:"display_name"`
	BaseURL          string     `gorm:"size:256" json:"base_url"`
	ListPath         string     `gorm:"size:256" json:"list_path"`
	DetailPathTpl    string     `gorm:"size:256" json:"detail_path_tpl"`
	Enabled          bool       `gorm:"default:true" json:"enabled"`
	IntervalSec      int        `gorm:"default:900" json:"interval_sec"`
	Concurrency      int        `gorm:"default:4" json:"concurrency"`
	MaxPagesPerCycle int        `gorm:"default:5" json:"max_pages_per_cycle"`
	LastRunAt        *time.Time `json:"last_run_at"`
	LastError        string     `gorm:"type:text" json:"last_error"`
	TypeMap          string     `gorm:"type:text" json:"type_map"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type Item struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	SourceID     uint64     `gorm:"index" json:"source_id"`
	ExternalID   string     `gorm:"index:idx_source_ext,unique,priority:2;size:128" json:"external_id"`
	URL          string     `gorm:"size:512" json:"url"`
	Title        string     `gorm:"type:text" json:"title"`
	TitleRaw     string     `gorm:"type:text" json:"title_raw"`
	TypeSlug     string     `gorm:"index;size:64" json:"type_slug"`
	TypeLabel    string     `gorm:"size:128" json:"type_label"`
	TypeDhivehi  string     `gorm:"size:128" json:"type_dhivehi"`
	Office       string     `gorm:"index;size:256" json:"office"`
	ReferenceNo  string     `gorm:"index;size:128" json:"reference_no"`
	PublishedAt  *time.Time `json:"published_at"`
	DeadlineAt   *time.Time `json:"deadline_at"`
	BodyText     string     `gorm:"type:text" json:"body_text"`
	BodyHTML     string     `gorm:"type:text" json:"body_html"`
	ContentHash  string     `gorm:"index;size:64" json:"content_hash"`
	FetchedAt    time.Time  `json:"fetched_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type FlagRule struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:128" json:"name"`
	Pattern       string    `gorm:"type:text" json:"pattern"`
	IsRegex       bool      `json:"is_regex"`
	MatchIn       string    `gorm:"size:16;default:'body'" json:"match_in"`
	CaseSensitive bool      `json:"case_sensitive"`
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FlagHit struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ItemID    uint64    `gorm:"index:idx_hit_item_rule,unique,priority:1" json:"item_id"`
	RuleID    uint64    `gorm:"index:idx_hit_item_rule,unique,priority:2" json:"rule_id"`
	Snippet   string    `gorm:"type:text" json:"snippet"`
	Notified  bool      `gorm:"default:false" json:"notified"`
	CreatedAt time.Time `json:"created_at"`
}

type TelegramConfig struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	BotToken         string    `gorm:"size:256" json:"bot_token"`
	Enabled          bool      `gorm:"default:false" json:"enabled"`
	NotifyOnFlagOnly bool      `gorm:"default:true" json:"notify_on_flag_only"`
	ThrottleMs       int       `gorm:"default:250" json:"throttle_ms"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type TelegramSubscriber struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ChatID    int64     `gorm:"uniqueIndex" json:"chat_id"`
	Label     string    `gorm:"size:128" json:"label"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type ScrapeRun struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	SourceID     uint64     `gorm:"index" json:"source_id"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	NewItems     int        `json:"new_items"`
	UpdatedItems int        `json:"updated_items"`
	Fetched      int        `json:"fetched"`
	Errors       int        `json:"errors"`
	ErrorLog     string     `gorm:"type:text" json:"error_log"`
}

type AppMeta struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

type UIPref struct {
	Key   string `gorm:"primaryKey;size:64" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (Item) TableName() string         { return "items" }
func (User) TableName() string         { return "users" }
func (Source) TableName() string       { return "sources" }
func (FlagRule) TableName() string     { return "flag_rules" }
func (FlagHit) TableName() string      { return "flag_hits" }
func (TelegramConfig) TableName() string {
	if false {
		_ = strings.TrimSpace
	}
	return "telegram_config"
}
func (TelegramSubscriber) TableName() string  { return "telegram_subscribers" }
func (ScrapeRun) TableName() string          { return "scrape_runs" }
func (AppMeta) TableName() string            { return "app_meta" }
func (UIPref) TableName() string             { return "ui_prefs" }
