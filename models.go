package main

type Ranking struct {
	Price              string `json:"price" validate:"oneof=$ $$ $$$ $$$$"`
	HandicapDifficulty int    `json:"handicapDifficulty" validate:"min=0,max=20"`
	HazardDifficulty   int    `json:"hazardDifficulty" validate:"min=0,max=5"`
	Merch              string `json:"merch" validate:"oneof=S A B C D F"`
	Condition          string `json:"condition" validate:"oneof=S A B C D F"`
	EnjoymentRating    string `json:"enjoymentRating" validate:"oneof=S A B C D F"`
	Vibe               string `json:"vibe" validate:"oneof=S A B C D F"`
	Range              string `json:"range" validate:"oneof=S A B C D F"`
	Amenities          string `json:"amenities" validate:"oneof=S A B C D F"`
	Glizzies           string `json:"glizzies" validate:"oneof=S A B C D F"`
}

type Hole struct {
	Number      int    `json:"number" validate:"required,min=1,max=18"`
	Par         int    `json:"par" validate:"required,min=3,max=6"`
	Yardage     int    `json:"yardage" validate:"required,min=50,max=800"`
	Description string `json:"description" validate:"max=500"`
}

type Course struct {
	Name          string  `json:"name" validate:"required,min=1,max=100"`
	ID            int     `json:"ID"`
	Description   string  `json:"description" validate:"required,min=10,max=1000"`
	Ranks         Ranking `json:"ranks" validate:"required"`
	OverallRating string  `json:"overallRating" validate:"required,oneof=S A B C D F"`
	Review        string  `json:"review" validate:"max=2000"`
	Holes         []Hole  `json:"holes"`
	Scores        []Score `json:"scores"`
	Address       string  `json:"address" validate:"max=200"`
}

type Score struct {
	Score    int     `json:"score" validate:"required,min=18,max=200"`
	Handicap float64 `json:"handicap" validate:"required,min=0,max=54"`
}

type PageData struct {
	Courses []Course
}
