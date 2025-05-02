package category

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/maximbilan/mcc"
)

var categories map[string][]string

func init() {
	categories = map[string][]string{
		"Huge":       {"Car", "Dwelling"},
		"Bills":      {"Utilities", "Cellurar", "Internet", "Other"},
		"Devices":    {},
		"Gifts":      {},
		"Car":        {"Accessories", "Insurance", "Garage", "Fuel", "Rent", "Maintenance", "Parking", "Other"},
		"Children":   {"Vocal", "Things", "Hospital", "Kindergarten", "Other"},
		"Business":   {"Broker", "Taxes", "Travel", "Accounts", "Software", "Translations", "Accountability", "Salary", "Design", "Lawyer", "Fee", "Finances", "Other"},
		"Help":       {"Donation", "Family", "Other"},
		"Transport":  {"Subway", "Taxi", "Bus", "Plane", "Train", "Other"},
		"Activities": {"Swimming", "Cinema", "Activities", "Sport", "Other", "F1"},
		"Food":       {"Shop", "Alcohol", "Outdoors", "Other"},
		"Things":     {"Clothes", "Shoes", "Accessories", "Other"},
		"Education":  {"Language", "Other"},
		"Health":     {"Mental", "Dentist", "Vision", "Pharmacy", "Medicine", "Other"},
		"House":      {"Furniture", "Maintenance", "Other"},
		"Multimedia": {"Applications", "Books", "Movies", "Music", "Storage", "Games", "Other"},
		"Travel":     {"Permission", "Hotel", "Excursion", "Other"},
		"Waste":      {},
		"Other":      {},
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
