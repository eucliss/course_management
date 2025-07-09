package api

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// MapHandler handles map and location-related API endpoints
type MapHandler struct {
	dbService MapDatabaseServiceInterface
}

// MapCourseResponse represents course data optimized for map display
type MapCourseResponse struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	// Minimal additional data for map markers
	OverallRating *float64 `json:"overall_rating,omitempty"`
	TotalReviews  int      `json:"total_reviews"`
	CanEdit       bool     `json:"can_edit"`
	Distance      *float64 `json:"distance,omitempty"` // in kilometers
}

// BoundsRequest represents geographic bounds for map queries
type BoundsRequest struct {
	NorthLat  float64 `query:"north_lat" validate:"required,min=-90,max=90"`
	SouthLat  float64 `query:"south_lat" validate:"required,min=-90,max=90"`
	EastLng   float64 `query:"east_lng" validate:"required,min=-180,max=180"`
	WestLng   float64 `query:"west_lng" validate:"required,min=-180,max=180"`
	MinRating *float64 `query:"min_rating" validate:"omitempty,min=0,max=10"`
	MaxRating *float64 `query:"max_rating" validate:"omitempty,min=0,max=10"`
}

// GeocodeRequest represents a geocoding request
type GeocodeRequest struct {
	Address string `json:"address" validate:"required,min=5,max=200"`
}

// GeocodeResponse represents geocoding results
type GeocodeResponse struct {
	Address         string   `json:"address"`
	FormattedAddress string  `json:"formatted_address"`
	Latitude        float64  `json:"latitude"`
	Longitude       float64  `json:"longitude"`
	Confidence      float64  `json:"confidence"` // 0.0 to 1.0
	Components      *AddressComponents `json:"components,omitempty"`
}

// AddressComponents represents parsed address components
type AddressComponents struct {
	StreetNumber string `json:"street_number,omitempty"`
	Route        string `json:"route,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	Country      string `json:"country,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
}

// CourseClusterResponse represents clustered course data for map display
type CourseClusterResponse struct {
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CourseCount int     `json:"course_count"`
	ZoomLevel   int     `json:"zoom_level"`
	Courses     []*MapCourseResponse `json:"courses,omitempty"` // Only included for small clusters
}

// NewMapHandler creates a new map handler
func NewMapHandler(dbService MapDatabaseServiceInterface) *MapHandler {
	return &MapHandler{
		dbService: dbService,
	}
}

// GetMapCourses returns courses optimized for map display
func (h *MapHandler) GetMapCourses(c echo.Context) error {
	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	// Get all courses with location data
	courses, err := h.dbService.GetMapCourses(userID)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve map courses")
	}

	return SuccessResponse(c, courses)
}

// GetCoursesInBounds returns courses within geographic bounds
func (h *MapHandler) GetCoursesInBounds(c echo.Context) error {
	var bounds BoundsRequest
	if err := c.Bind(&bounds); err != nil {
		return BadRequestError(c, "Invalid bounds parameters")
	}

	// Validate bounds
	if bounds.NorthLat <= bounds.SouthLat {
		return BadRequestError(c, "North latitude must be greater than south latitude")
	}

	if bounds.EastLng <= bounds.WestLng {
		return BadRequestError(c, "East longitude must be greater than west longitude")
	}

	// Validate rating range
	if bounds.MinRating != nil && bounds.MaxRating != nil && *bounds.MinRating > *bounds.MaxRating {
		return BadRequestError(c, "Minimum rating cannot be greater than maximum rating")
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	courses, err := h.dbService.GetCoursesInBounds(&bounds, userID)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve courses in bounds")
	}

	return SuccessResponse(c, courses)
}

// GetClusteredCourses returns clustered course data for efficient map rendering
func (h *MapHandler) GetClusteredCourses(c echo.Context) error {
	var bounds BoundsRequest
	if err := c.Bind(&bounds); err != nil {
		return BadRequestError(c, "Invalid bounds parameters")
	}

	// Get zoom level for clustering
	zoomLevelParam := c.QueryParam("zoom")
	zoomLevel := 10 // Default zoom level
	if zoomLevelParam != "" {
		if z, err := strconv.Atoi(zoomLevelParam); err == nil && z >= 1 && z <= 20 {
			zoomLevel = z
		}
	}

	// Get cluster size preference
	maxClusterSizeParam := c.QueryParam("max_cluster_size")
	maxClusterSize := 50 // Default max cluster size
	if maxClusterSizeParam != "" {
		if s, err := strconv.Atoi(maxClusterSizeParam); err == nil && s >= 10 && s <= 1000 {
			maxClusterSize = s
		}
	}

	// Get user ID if authenticated
	var userID *uint
	if uid, err := GetUserID(c); err == nil {
		userID = &uid
	}

	clusters, err := h.dbService.GetClusteredCourses(&bounds, userID, zoomLevel, maxClusterSize)
	if err != nil {
		return InternalServerError(c, "Failed to retrieve clustered courses")
	}

	return SuccessResponse(c, clusters)
}

// GeocodeAddress geocodes an address to coordinates
func (h *MapHandler) GeocodeAddress(c echo.Context) error {
	var req GeocodeRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestError(c, "Invalid request format")
	}

	if len(req.Address) < 5 {
		return ValidationError(c, map[string]string{
			"address": "Address must be at least 5 characters long",
		})
	}

	result, err := h.dbService.GeocodeAddress(req.Address)
	if err != nil {
		return InternalServerError(c, "Geocoding failed")
	}

	if result == nil {
		return NotFoundError(c, "Address could not be geocoded")
	}

	return SuccessResponse(c, result)
}

// ReverseGeocode converts coordinates to address
func (h *MapHandler) ReverseGeocode(c echo.Context) error {
	latParam := c.QueryParam("lat")
	lngParam := c.QueryParam("lng")

	if latParam == "" || lngParam == "" {
		return BadRequestError(c, "Latitude and longitude are required")
	}

	latitude, err := strconv.ParseFloat(latParam, 64)
	if err != nil || latitude < -90 || latitude > 90 {
		return BadRequestError(c, "Invalid latitude")
	}

	longitude, err := strconv.ParseFloat(lngParam, 64)
	if err != nil || longitude < -180 || longitude > 180 {
		return BadRequestError(c, "Invalid longitude")
	}

	result, err := h.dbService.ReverseGeocode(latitude, longitude)
	if err != nil {
		return InternalServerError(c, "Reverse geocoding failed")
	}

	if result == nil {
		return NotFoundError(c, "Coordinates could not be reverse geocoded")
	}

	return SuccessResponse(c, result)
}

// GetCourseRoute returns directions between two courses or from location to course
func (h *MapHandler) GetCourseRoute(c echo.Context) error {
	fromCourseIDParam := c.QueryParam("from_course_id")
	toCourseIDParam := c.QueryParam("to_course_id")
	fromLatParam := c.QueryParam("from_lat")
	fromLngParam := c.QueryParam("from_lng")
	toLatParam := c.QueryParam("to_lat")
	toLngParam := c.QueryParam("to_lng")

	var fromLat, fromLng, toLat, toLng float64
	var err error

	// Parse "from" location
	if fromCourseIDParam != "" {
		fromCourseID, err := strconv.ParseUint(fromCourseIDParam, 10, 32)
		if err != nil {
			return BadRequestError(c, "Invalid from_course_id")
		}
		
		course, err := h.dbService.GetCourseLocation(uint(fromCourseID))
		if err != nil {
			return NotFoundError(c, "From course not found")
		}
		
		if course.Latitude == nil || course.Longitude == nil {
			return BadRequestError(c, "From course does not have location data")
		}
		
		fromLat = *course.Latitude
		fromLng = *course.Longitude
	} else if fromLatParam != "" && fromLngParam != "" {
		fromLat, err = strconv.ParseFloat(fromLatParam, 64)
		if err != nil || fromLat < -90 || fromLat > 90 {
			return BadRequestError(c, "Invalid from_lat")
		}
		
		fromLng, err = strconv.ParseFloat(fromLngParam, 64)
		if err != nil || fromLng < -180 || fromLng > 180 {
			return BadRequestError(c, "Invalid from_lng")
		}
	} else {
		return BadRequestError(c, "Either from_course_id or from_lat/from_lng must be provided")
	}

	// Parse "to" location
	if toCourseIDParam != "" {
		toCourseID, err := strconv.ParseUint(toCourseIDParam, 10, 32)
		if err != nil {
			return BadRequestError(c, "Invalid to_course_id")
		}
		
		course, err := h.dbService.GetCourseLocation(uint(toCourseID))
		if err != nil {
			return NotFoundError(c, "To course not found")
		}
		
		if course.Latitude == nil || course.Longitude == nil {
			return BadRequestError(c, "To course does not have location data")
		}
		
		toLat = *course.Latitude
		toLng = *course.Longitude
	} else if toLatParam != "" && toLngParam != "" {
		toLat, err = strconv.ParseFloat(toLatParam, 64)
		if err != nil || toLat < -90 || toLat > 90 {
			return BadRequestError(c, "Invalid to_lat")
		}
		
		toLng, err = strconv.ParseFloat(toLngParam, 64)
		if err != nil || toLng < -180 || toLng > 180 {
			return BadRequestError(c, "Invalid to_lng")
		}
	} else {
		return BadRequestError(c, "Either to_course_id or to_lat/to_lng must be provided")
	}

	route, err := h.dbService.GetRoute(fromLat, fromLng, toLat, toLng)
	if err != nil {
		return InternalServerError(c, "Failed to calculate route")
	}

	return SuccessResponse(c, route)
}

// GetMapStatistics returns map-related statistics
func (h *MapHandler) GetMapStatistics(c echo.Context) error {
	stats, err := h.dbService.GetMapStatistics()
	if err != nil {
		return InternalServerError(c, "Failed to retrieve map statistics")
	}

	return SuccessResponse(c, stats)
}

// RegisterRoutes registers map-related routes
func (h *MapHandler) RegisterRoutes(g *echo.Group, jwtService *JWTService) {
	// Public routes (optionally authenticated)
	g.GET("/map/courses", h.GetMapCourses, OptionalJWTMiddleware(jwtService))
	g.GET("/map/courses/bounds", h.GetCoursesInBounds, OptionalJWTMiddleware(jwtService))
	g.GET("/map/courses/clusters", h.GetClusteredCourses, OptionalJWTMiddleware(jwtService))
	g.GET("/map/statistics", h.GetMapStatistics)
	
	// Geocoding routes (no authentication required)
	g.POST("/map/geocode", h.GeocodeAddress)
	g.GET("/map/reverse-geocode", h.ReverseGeocode)
	g.GET("/map/route", h.GetCourseRoute)
}

// Route response structures
type RouteResponse struct {
	Distance     float64          `json:"distance"`     // in kilometers
	Duration     int              `json:"duration"`     // in seconds
	Geometry     string           `json:"geometry"`     // encoded polyline
	Steps        []*RouteStep     `json:"steps"`
	Waypoints    []*RouteWaypoint `json:"waypoints"`
}

type RouteStep struct {
	Distance     float64  `json:"distance"`    // in meters
	Duration     int      `json:"duration"`    // in seconds
	Instruction  string   `json:"instruction"`
	Geometry     string   `json:"geometry"`
	StartLat     float64  `json:"start_lat"`
	StartLng     float64  `json:"start_lng"`
	EndLat       float64  `json:"end_lat"`
	EndLng       float64  `json:"end_lng"`
}

type RouteWaypoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
}

// Map statistics response
type MapStatisticsResponse struct {
	TotalCourses       int              `json:"total_courses"`
	CoursesWithLocation int             `json:"courses_with_location"`
	LocationCoverage   float64          `json:"location_coverage"` // percentage
	BoundingBox        *GeographicBounds `json:"bounding_box"`
	PopularRegions     []*RegionStats   `json:"popular_regions"`
}

type GeographicBounds struct {
	NorthLat float64 `json:"north_lat"`
	SouthLat float64 `json:"south_lat"`
	EastLng  float64 `json:"east_lng"`
	WestLng  float64 `json:"west_lng"`
}

type RegionStats struct {
	Name         string  `json:"name"`
	CourseCount  int     `json:"course_count"`
	AverageRating float64 `json:"average_rating"`
	CenterLat    float64 `json:"center_lat"`
	CenterLng    float64 `json:"center_lng"`
}

// Extended database interface for map operations
type MapDatabaseServiceInterface interface {
	ReviewDatabaseServiceInterface
	GetMapCourses(userID *uint) ([]*MapCourseResponse, error)
	GetCoursesInBounds(bounds *BoundsRequest, userID *uint) ([]*MapCourseResponse, error)
	GetClusteredCourses(bounds *BoundsRequest, userID *uint, zoomLevel, maxClusterSize int) ([]*CourseClusterResponse, error)
	GeocodeAddress(address string) (*GeocodeResponse, error)
	ReverseGeocode(lat, lng float64) (*GeocodeResponse, error)
	GetCourseLocation(courseID uint) (*MapCourseResponse, error)
	GetRoute(fromLat, fromLng, toLat, toLng float64) (*RouteResponse, error)
	GetMapStatistics() (*MapStatisticsResponse, error)
}