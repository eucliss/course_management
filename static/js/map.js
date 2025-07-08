// Shared map functionality for HTMX app
window.CourseMap = {
    map: null,
    initialized: false,
    
    init(containerId, mapboxToken, courses = []) {
        console.log('🗺️ Initializing map in container:', containerId);
        
        if (!mapboxToken) {
            console.error('❌ Mapbox token missing');
            return;
        }
        
        mapboxgl.accessToken = mapboxToken;
        
        try {
            this.map = new mapboxgl.Map({
                container: containerId,
                style: 'mapbox://styles/mapbox/streets-v12',
                center: [-98.5795, 39.8283],
                zoom: 4.5,
                minZoom: 3,
                maxZoom: 15,
                renderWorldCopies: false,
                antialias: false,
                preserveDrawingBuffer: false
            });
            
            this.map.on('error', (e) => {
                console.error('🗺️ Map error:', e);
            });
            
            this.map.on('load', () => {
                console.log('✅ Map loaded successfully');
                this.loadCourses(courses);
                this.initialized = true;
            });
            
        } catch (error) {
            console.error('💥 Failed to initialize map:', error);
            document.getElementById(containerId).innerHTML = '<p>Failed to load map. Check console for errors.</p>';
        }
    },
    
    loadCourses(courses) {
        console.log('🏌️ Loading courses on map:', courses.length);
        
        if (!courses || courses.length === 0) {
            console.log('ℹ️ No courses to display');
            return;
        }
        
        courses.forEach(course => {
            console.log('📍 Processing course:', course.Name, 'ID:', course.ID);
            
            if (course.Address && course.Address.trim() !== '') {
                this.addPin(course);
            } else {
                console.log('⚠️ Skipping course without address:', course.Name);
            }
        });
    },
    
    addPin(course) {
        const ratingColors = {
            'S': '#73FF73',
            'A': '#B7FF73', 
            'B': '#FFFF73',
            'C': '#FFDA74',
            'D': '#FFB774',
            'F': '#FF7474'
        };
        
        const markerColor = ratingColors[course.OverallRating] || '#204606';
        
        // Check if course has stored coordinates
        if (course.Latitude && course.Longitude) {
            // Use stored coordinates directly - much faster!
            console.log('📍 Using stored coordinates for:', course.Name, { lat: course.Latitude, lng: course.Longitude });
            
            const marker = new mapboxgl.Marker({
                color: markerColor,
                scale: 1.2
            })
            .setLngLat([course.Longitude, course.Latitude])
            .setPopup(
                new mapboxgl.Popup({ offset: 25 })
                    .setHTML(`
                        <div class="marker-popup">
                            <div class="popup-content">
                                <div class="popup-header">
                                    <h3 class="course-title" 
                                       data-course-id="${course.ID}" 
                                       style="cursor: pointer; text-decoration: underline;">
                                       ${course.Name}
                                    </h3>
                                    <p>Click to view course details</p>
                                </div>
                            </div>
                            <div class="rating-plaque" style="background-color: ${markerColor};">${course.OverallRating}</div>
                        </div>
                    `)
            )
            .addTo(this.map);

            // HTMX-compatible click handler
            marker.getPopup().on('open', () => {
                setTimeout(() => {
                    const titleElement = document.querySelector(`[data-course-id="${course.ID}"]`);
                    if (titleElement) {
                        titleElement.addEventListener('click', function() {
                            const courseId = this.getAttribute('data-course-id');
                            console.log('🎯 Course clicked:', courseId);
                            
                            // Use HTMX to load course details
                            htmx.ajax('GET', `/course/${courseId}`, {
                                target: '#main-content'
                            });
                        });
                    }
                }, 100);
            });
            
            return; // Exit early - we used stored coordinates
        }
        
        // FALLBACK: Use geocoding API if no stored coordinates and address exists
        if (!course.Address || course.Address.trim() === '') {
            console.log('⚠️ No address or coordinates for course:', course.Name);
            return;
        }
        
        
        fetch(`https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(course.Address)}.json?access_token=${mapboxgl.accessToken}`)
            .then(response => response.json())
            .then(data => {
                if (data.features && data.features.length > 0) {
                    const [lng, lat] = data.features[0].center;
                    console.log('📍 Geocoded coordinates for:', course.Name, { lat, lng });
                    
                    const marker = new mapboxgl.Marker({
                        color: markerColor,
                        scale: 1.2
                    })
                    .setLngLat([lng, lat])
                    .setPopup(
                        new mapboxgl.Popup({ offset: 25 })
                            .setHTML(`
                                <div class="marker-popup">
                                    <div class="popup-content">
                                        <div class="popup-header">
                                            <h3 class="course-title" 
                                               data-course-id="${course.ID}" 
                                               style="cursor: pointer; text-decoration: underline;">
                                               ${course.Name}
                                            </h3>
                                            <p>Click to view course details</p>
                                        </div>
                                    </div>
                                    <div class="rating-plaque" style="background-color: ${markerColor};">${course.OverallRating}</div>
                                </div>
                            `)
                    )
                    .addTo(this.map);

                    // HTMX-compatible click handler
                    marker.getPopup().on('open', () => {
                        setTimeout(() => {
                            const titleElement = document.querySelector(`[data-course-id="${course.ID}"]`);
                            if (titleElement) {
                                titleElement.addEventListener('click', function() {
                                    const courseId = this.getAttribute('data-course-id');
                                    console.log('🎯 Course clicked:', courseId);
                                    
                                    // Use HTMX to load course details
                                    htmx.ajax('GET', `/course/${courseId}`, {
                                        target: '#main-content'
                                    });
                                });
                            }
                        }, 100);
                    });
                } else {
                    console.log('❌ Could not geocode:', course.Address);
                }
            })
            .catch(error => {
                console.error('💥 Geocoding error for', course.Name, ':', error);
            });
    },
    
    destroy() {
        if (this.map) {
            try {
                // Check if canvas exists before removing
                if (this.map.getCanvas()) {
                    this.map.remove();
                    console.log('🗑️ Map destroyed');
                }
            } catch (e) {
                console.log('🗑️ Error destroying map or canvas already gone:', e);
            } finally {
                this.map = null;
                this.initialized = false;
            }
        }
    }
};

// Enhanced HTMX integration for map cleanup
document.addEventListener('htmx:beforeSwap', function(e) {
    // Clean up map when navigating away from map page
    if (window.CourseMap && window.CourseMap.initialized) {
        window.CourseMap.destroy();
    }
    
    // Also clean up dedicated map instances
    if (window.currentMapInstance) {
        try {
            if (!window.currentMapInstance._removed && window.currentMapInstance.getCanvas()) {
                window.currentMapInstance.remove();
            }
        } catch (e) {
            console.log('🗑️ Error cleaning up current map instance:', e);
        }
        window.currentMapInstance = null;
    }
});

// Additional cleanup for HTMX navigation
document.addEventListener('htmx:beforeRequest', function(e) {
    // Only cleanup if we're navigating away from a map page
    const currentUrl = window.location.pathname;
    const targetUrl = e.detail.requestConfig.url || e.detail.requestConfig.path;
    
    if (currentUrl.includes('/map') && targetUrl && !targetUrl.includes('/map')) {
        console.log('🚀 Navigating away from map page, cleaning up...');
        
        // Clean up shared map
        if (window.CourseMap && window.CourseMap.initialized) {
            window.CourseMap.destroy();
        }
        
        // Clean up dedicated map instances
        if (window.currentMapInstance) {
            try {
                if (!window.currentMapInstance._removed && window.currentMapInstance.getCanvas()) {
                    window.currentMapInstance.remove();
                }
            } catch (e) {
                console.log('🗑️ Error cleaning up during navigation:', e);
            }
            window.currentMapInstance = null;
        }
    }
}); 