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
        
        // Escape data before using in popup
        const escapedName = this.escapeHtml(course.Name);
        const escapedRating = this.escapeHtml(course.OverallRating);
        
        const popup = new mapboxgl.Popup({ offset: 25 })
            .setHTML(`
                <div class="marker-popup">
                    <div class="popup-content">
                        <div class="popup-header">
                            <h3 class="course-title" data-course-id="${course.ID}">${escapedName}</h3>
                            <p>Click to view course details</p>
                        </div>
                    </div>
                    <div class="rating-plaque">${escapedRating}</div>
                </div>
            `);
        
        fetch(`https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(course.Address)}.json?access_token=${mapboxgl.accessToken}`)
            .then(response => response.json())
            .then(data => {
                if (data.features && data.features.length > 0) {
                    const [lng, lat] = data.features[0].center;
                    console.log('📍 Geocoded:', course.Name, { lat, lng });
                    
                    const marker = new mapboxgl.Marker({
                        color: markerColor,
                        scale: 1.2
                    })
                    .setLngLat([lng, lat])
                    .setPopup(popup)
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
    
    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    },
    
    destroy() {
        if (this.map) {
            this.map.remove();
            this.map = null;
            this.initialized = false;
            console.log('🗑️ Map destroyed');
        }
    }
};

// HTMX integration
document.addEventListener('htmx:beforeSwap', function(e) {
    // Clean up map when navigating away
    if (window.CourseMap && window.CourseMap.initialized) {
        window.CourseMap.destroy();
    }
}); 