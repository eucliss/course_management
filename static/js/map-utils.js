// Shared map utility functions
window.MapUtils = {
  // Rating color definitions
  ratingColors: {
    'S': '#73FF73',
    'A': '#B7FF73',
    'B': '#FFFF73',
    'C': '#FFDA74',
    'D': '#FFB774',
    'F': '#FF7474',
    '-': '#BABEBC'
  },

  // Create enhanced popup HTML
  createEnhancedPopupHTML: function(courseName, courseId, rating, address) {
    return `
      <div class="popup-header">
        <h3 class="popup-title" data-course-id="${courseId}">${courseName}</h3>
        <div class="popup-rating-badge rating-${rating === '-' ? 'none' : rating}">${rating || '-'}</div>
      </div>
      ${address ? `<p class="popup-address">${address}</p>` : ''}
    `;
  },

  // Create basic popup HTML
  createPopupHTML: function(courseName, courseId, rating, markerColor) {
    return `
      <div class="marker-popup">
        <div class="popup-content">
          <div class="popup-header">
            <h3 class="course-title" data-course-id="${courseId}">${courseName}</h3>
          </div>
        </div>
        <div class="rating-plaque" style="background-color: ${markerColor};">${rating || '-'}</div>
      </div>
    `;
  },

  // Add click handler to popup titles
  addPopupClickHandler: function(delay = 100) {
    setTimeout(() => {
      const titleElements = document.querySelectorAll('.popup-title, .course-title');
      const titleElement = titleElements[titleElements.length - 1];
      if (titleElement) {
        titleElement.addEventListener('click', function() {
          const courseId = this.getAttribute('data-course-id');
          console.log('🎯 Course clicked:', courseId);
          
          if (courseId === null || courseId === undefined || courseId === 'undefined') {
            console.error('❌ Invalid course ID:', courseId);
            return;
          }
          
          const url = `/course/${courseId}`;
          console.log('🚀 Navigating to:', url);
          
          htmx.ajax('GET', url, {
            target: '#main-content'
          });
        });
      }
    }, delay);
  },

  // Get marker color based on rating
  getMarkerColor: function(rating) {
    return this.ratingColors[rating] || this.ratingColors['-'];
  },

  // Create GeoJSON feature from course data
  createGeoJSONFeature: function(course, index, editPermissions, reviewStatus, userReviewMap, currentFilter) {
    const courseName = course.Name || course.name;
    const hasReview = reviewStatus && reviewStatus[courseName] === true;
    
    // Skip if filtering for "my courses" and user hasn't reviewed this course
    if (currentFilter === 'my' && !hasReview) {
      return null;
    }
    
    // Determine rating based on filter
    let rating = course.OverallRating || course.overallRating || '-';
    if (currentFilter === 'my' && userReviewMap[courseName]) {
      const userReview = userReviewMap[courseName];
      rating = userReview.OverallRating || userReview.overallRating || rating;
    }
    
    const courseId = course.ID !== undefined ? course.ID : index;
    
    return {
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: [course.longitude, course.latitude]
      },
      properties: {
        name: courseName,
        address: course.Address || course.address || '',
        rating: rating,
        color: this.getMarkerColor(rating),
        courseId: courseId,
        canEdit: editPermissions && editPermissions[courseName] === true,
        hasReview: hasReview
      }
    };
  },

  // Clean up existing map instance
  cleanupMap: function(mapInstance) {
    if (mapInstance) {
      try {
        if (!mapInstance._removed && mapInstance.getCanvas()) {
          console.log('🔄 Cleaning up existing map instance');
          mapInstance.remove();
        }
      } catch (e) {
        console.log('🗑️ Map already removed or error during cleanup:', e);
      }
    }
  },

  // Clean up markers array
  cleanupMarkers: function(markersArray) {
    if (markersArray) {
      markersArray.forEach(markerData => {
        if (markerData.marker) {
          markerData.marker.remove();
        }
      });
    }
    return [];
  },

  // Add course from address (geocoding fallback)
  addCourseFromAddress: function(map, address, courseName, rating, courseId, courseIndex, mapboxToken) {
    if (!address || address.trim() === '') {
      console.log('⚠️ No address for course:', courseName);
      return;
    }
    
    console.log('📍 Geocoding address for:', courseName);
    
    const markerColor = this.getMarkerColor(rating);
    
    return fetch(`https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(address)}.json?access_token=${mapboxToken}`)
      .then(response => response.json())
      .then(data => {
        if (data.features && data.features.length > 0) {
          const [lng, lat] = data.features[0].center;
          
          let marker = new mapboxgl.Marker({
            color: markerColor,
            scale: 1.2
          })
          .setLngLat([lng, lat])
          .setPopup(
            new mapboxgl.Popup({ offset: 25 })
              .setHTML(this.createPopupHTML(courseName, courseId, rating, markerColor))
          )
          .addTo(map);

          // Add click handler to popup
          marker.getPopup().on('open', () => {
            this.addPopupClickHandler();
          });

          return marker;
        } else {
          console.log('❌ Could not geocode:', address);
          return null;
        }
      })
      .catch(error => {
        console.error('💥 Geocoding error:', error);
        return null;
      });
  }
};