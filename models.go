package main

type Ranking struct {
	Price              string `json:"price"`
	HandicapDifficulty int    `json:"handicapDifficulty"`
	HazardDifficulty   int    `json:"hazardDifficulty"`
	Merch              string `json:"merch"`
	Condition          string `json:"condition"`
	EnjoymentRating    string `json:"enjoymentRating"`
	Vibe               string `json:"vibe"`
	Range              string `json:"range"`
	Amenities          string `json:"amenities"`
	Glizzies           string `json:"glizzies"`
}

type Hole struct {
	Number      int    `json:"number"`
	Par         int    `json:"par"`
	Yardage     int    `json:"yardage"`
	Description string `json:"description"`
}

type Course struct {
	Name          string   `json:"name"`
	ID            int      `json:"ID"`
	Description   string   `json:"description"`
	Ranks         Ranking  `json:"ranks"`
	OverallRating string   `json:"overallRating"`
	Review        string   `json:"review"`
	Holes         []Hole   `json:"holes"`
	Scores        []Score  `json:"scores"`
	Address       string   `json:"address"`
	Latitude      *float64 `json:"latitude"`  // Geocoded latitude
	Longitude     *float64 `json:"longitude"` // Geocoded longitude
}

type Score struct {
	Score    int     `json:"score"`
	Handicap float64 `json:"handicap"`
}

type PageData struct {
	Courses []Course
}
