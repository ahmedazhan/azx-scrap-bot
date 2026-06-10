package db

import (
	"encoding/json"
	"log/slog"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TypeMapEntry struct {
	Dhivehi string `json:"dhivehi"`
	English string `json:"english"`
}

func Seed(db *gorm.DB, log *slog.Logger) error {
	typeMap := map[string]TypeMapEntry{
		"masakkaiy":          {"މަސައްކަތް", "Vacancy"},
		"gannan-beynunvaa":   {"ގަންނަން ބޭނުންވާ ތަކެތި", "Operational Notice"},
		"kuyyah-dhinun":      {"ކުއްޔަށް ދިނުން", "Lost"},
		"kuyyah-hifun":       {"ކުއްޔަށް ހިފުން", "Found"},
		"vazeefaa":           {"ވަޒީފާގެ ފުރުޞަތު", "Job Posting"},
		"thamreenu":          {"ތަމްރީނު", "Notice"},
		"neelan":             {"ނީލަން", "Tender"},
		"aanmu-mauloomaathu": {"ޢާންމު މަޢުލޫމާތު", "Public Information"},
		"dhennevun":          {"ދެންނެވުން", "Announcement"},
		"mubaaraaiy":         {"މުބާރާތް", "Bid"},
		"noos-bayaan":        {"ނޫސްބަޔާން", "News"},
		"insurance":          {"އިންޝުއަރެންސް", "Insurance"},
		"beelan":             {"ބީލަން", "Bill"},
	}
	typeMapJSON, err := json.Marshal(typeMap)
	if err != nil {
		return err
	}

	src := Source{
		Key:              "gazette-iulaan",
		DisplayName:      "Maldives Gazette — Iulaan",
		BaseURL:          "https://gazette.gov.mv",
		ListPath:         "/iulaan",
		DetailPathTpl:    "/iulaan/{id}",
		Enabled:          true,
		IntervalSec:      900,
		Concurrency:      4,
		MaxPagesPerCycle: 5,
		TypeMap:          string(typeMapJSON),
	}

	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"display_name", "base_url", "list_path", "detail_path_tpl", "type_map", "updated_at"}),
	}).Create(&src).Error
}
