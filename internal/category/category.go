package category

import (
	"strconv"
	"strings"

	"github.com/maximbilan/mcc"
)

// order pins the sequence used in prompts and in the schema enum. Map iteration
// is random, and an enum that reorders every call would defeat prompt caching.
var order = []string{
	"Huge", "Bills", "Devices", "Gifts", "Car", "Children", "Business", "Help",
	"Transport", "Activities", "Food", "Things", "Education", "Health", "House",
	"Multimedia", "Travel", "Waste", "Other",
}

var categories map[string][]string
var hints map[string]string
var leafHints map[string]string

func init() {
	categories = map[string][]string{
		"Huge":       {"Car", "Dwelling", "Other"},
		"Bills":      {"Utilities", "Cellurar", "Internet", "Other"},
		"Devices":    {"Phone", "Laptop", "Playstation", "TV Set", "Accessories", "Other"},
		"Gifts":      {"Family", "Friends", "Other"},
		"Car":        {"Accessories", "Insurance", "Garage", "Fuel", "Rent", "Maintenance", "Parking", "Other"},
		"Children":   {"Vocal", "Things", "Hospital", "Kindergarten", "Toys", "Other"},
		"Business":   {"Broker", "Taxes", "Travel", "Accounts", "Software", "Translations", "Accountability", "Salary", "Design", "Lawyer", "Fee", "Finances", "Other"},
		"Help":       {"Donation", "Family", "Friends", "Other"},
		"Transport":  {"Subway", "Taxi", "Bus", "Plane", "Train", "Other"},
		"Activities": {"Swimming", "Cinema", "Activities", "Sport", "Other", "F1"},
		"Food":       {"Shop", "Alcohol", "Outdoors", "Other"},
		"Things":     {"Clothes", "Shoes", "Accessories", "Other"},
		"Education":  {"Language", "Courses", "Other"},
		"Health":     {"Mental", "Dentist", "Vision", "Pharmacy", "Medicine", "Other"},
		"House":      {"Furniture", "Maintenance", "Details", "Other"},
		"Multimedia": {"Applications", "Books", "Movies", "Music", "Storage", "Games", "Other"},
		"Travel":     {"Permission", "Hotel", "Excursion", "Other"},
		"Waste":      {},
		"Other":      {},
	}

	hints = map[string]string{
		"Huge":       "Rarely used, but can be used for big purchases like car or house",
		"Bills":      "Bills for house utilities, internet, cellular, etc.",
		"Devices":    "Devices like phone, laptop, playstation, tv set, etc.",
		"Gifts":      "Any gifts for family, friends, etc.",
		"Car":        "Car expenses like fuel, insurance, maintenance, etc.",
		"Children":   "Expenses for children like kindergarten, hospital etc.",
		"Business":   "Expenses for business like taxes, software, translations, etc.",
		"Help":       "Any help for family, friends, donataions, etc.",
		"Transport":  "Expenses for transport like taxi, subway, bus, etc.",
		"Activities": "Expenses for activities like swimming, cinema, park attractions, any outside activities, make up for wife, etc.",
		"Food":       "Expenses for food like groceries (shop), alcohol, outdoors (restaraunt, cafe), etc.",
		"Things":     "Expenses for things like clothes, shoes, accessories, etc.",
		"Education":  "Expenses for education like language courses, certificates, etc.",
		"Health":     "Expenses for health like dentist, vision, pharmacy, medicine, etc.",
		"House":      "Expenses for house like furniture, maintenance, etc.",
		"Multimedia": "Expenses for online multimedia like applications, books, movies, music, storage, games, etc. For example: Netflix, Spotify, etc.",
		"Travel":     "Expenses for any travel things like permission (VISA), hotel, excursion, etc.",
		"Waste":      "Meaning I don't care about this expense",
		"Other":      "Any other expenses that don't fit into any category",
	}

	// Without per-leaf descriptions a small model cannot tell Business/Accounts
	// from Business/Finances, and falls back to ".../Other".
	leafHints = map[string]string{
		"Bills/Utilities":         "electricity, gas, water, heating, building or community fees",
		"Bills/Cellurar":          "mobile phone plans and top-ups",
		"Bills/Internet":          "home internet and TV provider",
		"Car/Fuel":                "petrol, diesel or charging at a filling station",
		"Car/Parking":             "parking lots, meters, toll roads",
		"Car/Garage":              "a rented garage or permanent parking place",
		"Car/Maintenance":         "service, repairs, tyres, car wash, inspection",
		"Car/Insurance":           "car insurance policies",
		"Car/Rent":                "renting a car",
		"Car/Accessories":         "car parts and gadgets bought separately",
		"Food/Shop":               "groceries: supermarkets, markets, bakeries, convenience stores",
		"Food/Outdoors":           "eating out: restaurants, cafes, bars, fast food, ice cream, coffee shops, food stalls",
		"Food/Alcohol":            "alcohol bought to take away",
		"Transport/Taxi":          "taxi and ride hailing (Uber, Bolt, Uklon)",
		"Transport/Subway":        "metro, tram and other local public transport",
		"Transport/Bus":           "city and intercity buses",
		"Transport/Plane":         "flight tickets and airline fees",
		"Transport/Train":         "train tickets",
		"Activities/Swimming":     "pool entry, swim club, swim gear",
		"Activities/Cinema":       "cinema tickets",
		"Activities/Activities":   "paid leisure outings: theme parks, zoos, museums, boat rides, tourist trains, attractions, playgrounds",
		"Activities/Sport":        "gym, sport events, sport equipment, ski passes and lifts",
		"Activities/F1":           "Formula 1 related spending",
		"Things/Clothes":          "clothing",
		"Things/Shoes":            "footwear",
		"Things/Accessories":      "bags, watches, jewellery, cosmetics, small personal items",
		"Health/Pharmacy":         "pharmacy purchases",
		"Health/Medicine":         "doctors, clinics, lab tests, treatments",
		"Health/Dentist":          "dental care",
		"Health/Vision":           "optician, glasses, lenses",
		"Health/Mental":           "therapy and psychologist",
		"House/Furniture":         "furniture",
		"House/Maintenance":       "repairs, tools, building materials, cleaning services",
		"House/Details":           "small household goods: decor, kitchenware, textiles, light bulbs, batteries",
		"Multimedia/Applications": "app store purchases and app subscriptions",
		"Multimedia/Storage":      "cloud storage such as iCloud, Google One, Dropbox",
		"Multimedia/Music":        "Spotify, Apple Music and similar",
		"Multimedia/Movies":       "Netflix, film rentals, streaming",
		"Multimedia/Books":        "books, e-books, audiobooks",
		"Multimedia/Games":        "video games and in-game purchases",
		"Travel/Hotel":            "hotels, apartments and hostels while travelling",
		"Travel/Excursion":        "guided tours and attractions while travelling abroad",
		"Travel/Permission":       "visas, travel permits, travel insurance",
		"Business/Taxes":          "taxes and mandatory state contributions; a Ukrainian treasury payee such as ГУК or Казначейство belongs here",
		"Business/Accounts":       "moving money between my own accounts, including transfers to and from the ФОП accounts and between my own cards",
		"Business/Finances":       "banking, currency exchange, cashback withdrawal, investment moves",
		"Business/Fee":            "bank and payment-processing fees and commissions",
		"Business/Salary":         "salary paid out or received, including payouts from employer-of-record platforms",
		"Business/Software":       "business software and SaaS subscriptions",
		"Business/Broker":         "brokerage and investment platform activity",
		"Business/Accountability": "accountant and bookkeeping services",
		"Business/Lawyer":         "legal services",
		"Business/Translations":   "translation and notary services",
		"Business/Design":         "design and creative contractor work",
		"Business/Travel":         "business trips",
		"Gifts/Family":            "presents and money for family",
		"Gifts/Friends":           "presents and money for friends",
		"Help/Family":             "financial help to family",
		"Help/Friends":            "financial help to friends",
		"Help/Donation":           "charity and donations",
		"Children/Kindergarten":   "kindergarten and school fees",
		"Children/Toys":           "toys",
		"Children/Things":         "clothes and gear for children",
		"Children/Hospital":       "children's medical care",
		"Children/Vocal":          "singing and music lessons",
		"Education/Language":      "language lessons",
		"Education/Courses":       "courses, certifications, training",
		"Devices/Phone":           "phones",
		"Devices/Laptop":          "laptops and computers",
		"Devices/TV Set":          "televisions",
		"Devices/Playstation":     "game consoles",
		"Devices/Accessories":     "chargers, cables, headphones, peripherals",
		"Huge/Car":                "buying a car",
		"Huge/Dwelling":           "buying or renting property",
	}
}

// Paths returns every leaf as "Category/Subcategory", or bare "Category" for the
// two that have no subcategories. Stable across calls.
func Paths() []string {
	paths := make([]string, 0, 128)
	for _, name := range order {
		subs := categories[name]
		if len(subs) == 0 {
			paths = append(paths, name)
			continue
		}
		for _, sub := range subs {
			paths = append(paths, name+"/"+sub)
		}
	}
	return paths
}

// SplitPath splits a path into its two parts, discarding anything outside the
// taxonomy so a bad model answer never reaches the deep link.
func SplitPath(path string) (string, string) {
	name, sub, _ := strings.Cut(strings.TrimSpace(path), "/")
	return Normalize(name, sub)
}

// Normalize keeps only pairs that exist in the taxonomy. An unknown category
// becomes "Other"; an unknown subcategory falls back to that category's "Other".
func Normalize(name string, sub string) (string, string) {
	subs, ok := categories[name]
	if !ok {
		return "Other", ""
	}
	if len(subs) == 0 {
		return name, ""
	}
	for _, candidate := range subs {
		if candidate == sub {
			return name, sub
		}
	}
	for _, candidate := range subs {
		if candidate == "Other" {
			return name, "Other"
		}
	}
	return name, ""
}

// ClassificationPrompt is the system prompt shared by every entry point. It
// lists the taxonomy leaf by leaf and pushes the model off the "Other" leaves.
func ClassificationPrompt() string {
	var b strings.Builder
	b.WriteString(`You are a personal-finance transaction classifier for one specific user.
You are given one bank transaction or one bank push notification. Assign it exactly one path from the taxonomy below.

How to decide:
1. First work out what the counterparty really is. Use the merchant name, the MCC code and your knowledge of real businesses, including local ones. The user lives between Ukraine and Spain, so Ukrainian, Spanish and English merchant names are all common.
2. Then pick the single most likely leaf, and commit to it even when you are not certain.
3. "<Category>/Other" and the bare "Other" category are LAST RESORTS. Never pick them because you are unsure - pick them only when no leaf in the taxonomy could plausibly describe the transaction. When two leaves both fit, choose the more likely one.
4. The MCC is a strong hint but not the answer: a filling station MCC is Car/Fuel even when the merchant name sounds like something else, and a restaurant with a generic MCC is still Food/Outdoors.
5. Ukrainian and Russian merchant text is common. "ГУК" and "Казначейство" are the state treasury. "ФОП" is the user's own sole-proprietor business. "Переказ", "На картку" and "З картки" mean a transfer.

Taxonomy:
`)
	for _, name := range order {
		b.WriteString("- ")
		b.WriteString(name)
		if hint := hints[name]; hint != "" {
			b.WriteString(" - ")
			b.WriteString(hint)
		}
		b.WriteString("\n")
		for _, sub := range categories[name] {
			path := name + "/" + sub
			b.WriteString("    - ")
			b.WriteString(path)
			if hint := leafHints[path]; hint != "" {
				b.WriteString(" - ")
				b.WriteString(hint)
			}
			b.WriteString("\n")
		}
	}
	b.WriteString(`
Fill "merchant" first with a short description of what the counterparty is, then the remaining fields.`)
	return b.String()
}

func getCodeAsString(code int32) string {
	return strconv.Itoa(int(code))
}

func GetCategoryFromMCC(code int32) (string, error) {
	category, err := mcc.GetCategory(getCodeAsString(code))
	if err != nil {
		return "", err
	}
	return category, nil
}
