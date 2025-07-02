package main

// Add these methods to CourseService

/*
func (cs *CourseService) LoadCoursesOptimized(offset, limit int) ([]Course, int64, error) {
	if cs.useDB && cs.dbService != nil {
		return cs.dbService.GetCoursesWithPagination(offset, limit, false)
	}

	// Fallback to JSON with simulated pagination
	allCourses, err := cs.LoadCoursesFromJSON()
	if err != nil {
		return nil, 0, err
	}

	start := offset
	end := offset + limit
	if start >= len(allCourses) {
		return []Course{}, int64(len(allCourses)), nil
	}
	if end > len(allCourses) {
		end = len(allCourses)
	}

	return allCourses[start:end], int64(len(allCourses)), nil
}
*/
