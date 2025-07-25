package category

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/maximbilan/mcc"
)

var categories map[string][]string
var hints map[string]string

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
}

func GetCategoriesInJSON() string {
	jsonData, err := json.MarshalIndent(categories, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling categories to JSON:", err)
		return "{}"
	}
	return string(jsonData)
}

func GetHintsInJSON() string {
	jsonData, err := json.MarshalIndent(hints, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling hints to JSON:", err)
		return "{}"
	}
	return string(jsonData)
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
