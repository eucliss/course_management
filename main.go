package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseFiles(
			"views/welcome.html",
			"views/course.html",
			"views/introduction.html",
		)),
	}
}

type Block struct {
	Id int
}

type Blocks struct {
	Start  int
	Next   int
	More   bool
	Blocks []Block
}

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

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
}

func main() {
	// Get port from environment variable, default to 42069 if not set

	courses := []Course{
		{
			Name:          "Bath Country Club",
			ID:            0,
			Description:   "I'm parring Bath this year I can feel it.",
			OverallRating: "S",
			Ranks: Ranking{
				Price:              "$",
				HandicapDifficulty: 15,
				HazardDifficulty:   2,
				Condition:          "B",
				Merch:              "B",
				EnjoymentRating:    "A",
				Vibe:               "S",
				Range:              "C",
				Amenities:          "B",
				Glizzies:           "F",
			},
			Review: `The 1779 location of the Battle of Stono ferry - you'll probably have a couple fallen soldiers (balls) as well while you're playing around the river and all the ponds on the grounds. Pretty standard front nine through the neighborhoods and around the horse stable opens up to the signature back nine where the cannon and the river sit waiting. Having a few bad shots around the river and you may be looking at a 3 sleeve round - dont get too close to the ponds while looking for balls, the alligators might be lurking. Putting surface is pretty standard shapes with bermudagrass - one more slap in the face on 18 with the newer island green, hopefully you have a few more balls left to lose in this one last water feature. Sand beaches were nice and fluffy after the morning moisture burned off, some are weirdly shaped depending on the hole which adds a little spice to the mix.
Generally a really pleasant experience and the course is well maintained. 

Vibe feels local, but slightly elevated. Its best to look nice here and have your manners ready - Stono Ferry feels like a respectable club and they're definitely trying. Its semi-private though so the last group I was with didnt hesitate to rip 100 balls into the water and crack a few beers on hole 3 at 8:30 in the morning on a Thursday.`,
			Holes: []Hole{
				{Number: 1, Par: 4, Yardage: 350, Description: "This is a great hole."},
				{Number: 2, Par: 3, Yardage: 150, Description: "This is a great hole."},
				{Number: 3, Par: 5, Yardage: 450, Description: "This is a great hole."},
			},
			Scores: []Score{
				{Score: 72, Handicap: 6.4},
				{Score: 73, Handicap: 6.4},
			},
		},
		{
			Name:          "Kiawah Island Ocean Course",
			ID:            1,
			Description:   "That wind is no joke.",
			OverallRating: "A",
			Ranks: Ranking{
				Price:              "$$$$",
				HandicapDifficulty: 18,
				HazardDifficulty:   5,
				Condition:          "A",
				Merch:              "S",
				EnjoymentRating:    "S",
				Vibe:               "S",
				Range:              "S",
				Amenities:          "S",
				Glizzies:           "A",
			},
			Review: `The wind is no joke here.`,
		},
		{
			Name:          "Pinehurst No. 2",
			ID:            2,
			Description:   "Rolled off again.",
			OverallRating: "A",
			Ranks: Ranking{
				Price:              "$$$",
				Condition:          "A",
				HandicapDifficulty: 17,
				HazardDifficulty:   5,
				Merch:              "S",
				EnjoymentRating:    "S",
				Vibe:               "S",
				Range:              "S",
				Amenities:          "S",
				Glizzies:           "S",
			},
			Review: `The wind is no joke here.`,
		},
	}

	e := echo.New()
	e.Renderer = NewTemplates()
	e.Use(middleware.Logger())

	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "welcome", PageData{
			Courses: courses,
		})
	})

	e.GET("/introduction", func(c echo.Context) error {
		return c.Render(http.StatusOK, "introduction", PageData{
			Courses: courses,
		})
	})

	e.GET("/course/:id", func(c echo.Context) error {
		id := c.Param("id")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			return c.Render(http.StatusNotFound, "error", "Invalid course ID")
		}
		return c.Render(http.StatusOK, "course", courses[idInt])
	})

	e.Logger.Fatal(e.Start(":" + "8080"))
}

// e.GET("/blocks", func(c echo.Context) error {
//     startStr := c.QueryParam("start")
//     start, err := strconv.Atoi(startStr)
//     if err != nil {
//         start = 0
//     }

//     blocks := []Block{}
//     for i := start; i < start+10; i++ {
//         blocks = append(blocks, Block{Id: i})
//     }

//     template := "blocks"
//     if start == 0 {
//         template = "blocks-index"
//     }
//     return c.Render(http.StatusOK, template, Blocks{
//         Start:  start,
//         Next:   start + 10,
//         More:   start+10 < 100,
//         Blocks: blocks,
//     })
// })
