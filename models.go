package main

type Ranking struct {
	Price              string
	HandicapDifficulty int
	HazardDifficulty   int
	Merch              string
	Condition          string
	EnjoymentRating    string
	Vibe               string
	Range              string
	Amenities          string
	Glizzies           string
}

type Hole struct {
	Number      int
	Par         int
	Yardage     int
	Description string
}

type Course struct {
	Name          string
	ID            int
	Description   string
	Ranks         Ranking
	OverallRating string
	Review        string
	Holes         []Hole
	Scores        []Score
}

type Score struct {
	Score    int
	Handicap float64
}

type PageData struct {
	Courses []Course
}
