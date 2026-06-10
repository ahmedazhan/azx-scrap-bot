package translate

type Entry struct {
	Dhivehi string
	English string
}

var GazetteIulaan = map[string]Entry{
	"masakkaiy":           {"މަސައްކަތް", "Vacancy"},
	"gannan-beynunvaa":    {"ގަންނަން ބޭނުންވާ ތަކެތި", "Operational Notice"},
	"kuyyah-dhinun":       {"ކުއްޔަށް ދިނުން", "Lost"},
	"kuyyah-hifun":        {"ކުއްޔަށް ހިފުން", "Found"},
	"vazeefaa":            {"ވަޒީފާގެ ފުރުޞަތު", "Job Posting"},
	"thamreenu":           {"ތަމްރީނު", "Notice"},
	"neelan":              {"ނީލަން", "Tender"},
	"aanmu-mauloomaathu":  {"ޢާންމު މަޢުލޫމާތު", "Public Information"},
	"dhennevun":           {"ދެންނެވުން", "Announcement"},
	"mubaaraaiy":          {"މުބާރާތް", "Bid"},
	"noos-bayaan":         {"ނޫސްބަޔާން", "News"},
	"insurance":           {"އިންޝުއަރެންސް", "Insurance"},
	"beelan":              {"ބީލަން", "Bill"},
}

func Lookup(slug string) (Entry, bool) {
	e, ok := GazetteIulaan[slug]
	return e, ok
}

func Dhivehi(slug string) string {
	if e, ok := GazetteIulaan[slug]; ok {
		return e.Dhivehi
	}
	return slug
}

func English(slug string) string {
	if e, ok := GazetteIulaan[slug]; ok {
		return e.English
	}
	return slug
}
