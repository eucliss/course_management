#!/usr/bin/env python3
"""
Optimized script to scrape golf course data from GolfNow
"""

import requests
from bs4 import BeautifulSoup
from urllib.parse import urljoin
import time
import re
import json
import concurrent.futures
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry
import os
from typing import List, Dict, Optional, Tuple

class OptimizedGolfScraper:
    def __init__(self, max_workers=5, delay=0.5, timeout=15):
        self.max_workers = max_workers
        self.delay = delay
        self.timeout = timeout
        self.session = self._create_session()
        
    def _create_session(self):
        """Create optimized requests session with connection pooling and retries"""
        session = requests.Session()
        
        # Set up retry strategy
        retry_strategy = Retry(
            total=3,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504],
        )
        
        # Mount adapter with retry strategy
        adapter = HTTPAdapter(max_retries=retry_strategy, pool_connections=20, pool_maxsize=20)
        session.mount("http://", adapter)
        session.mount("https://", adapter)
        
        # Set headers
        session.headers.update({
            'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36',
            'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
            'Accept-Language': 'en-US,en;q=0.5',
            'Accept-Encoding': 'gzip, deflate',
            'Connection': 'keep-alive',
            'Upgrade-Insecure-Requests': '1',
        })
        
        return session
    
    def get_html(self, url: str) -> Optional[str]:
        """Optimized HTML fetching with error handling"""
        try:
            response = self.session.get(url, timeout=self.timeout)
            response.raise_for_status()
            return response.text
        except requests.RequestException as e:
            print(f"Error fetching {url}: {e}")
            return None
    
    def extract_destination_links(self, html: str, base_url: str) -> List[Tuple[str, str]]:
        """Extract destination links with optimized parsing"""
        soup = BeautifulSoup(html, 'lxml')  # lxml is faster than html.parser
        
        # More specific selector for better performance
        destination_wrappers = soup.select('.us-destination-wrapper.white.rounded')
        
        links = []
        for wrapper in destination_wrappers:
            for link in wrapper.select('a[href]'):  # More efficient selector
                link_text = link.get_text(strip=True)
                href = link.get('href')
                
                if link_text and href:
                    full_url = urljoin(base_url, href)
                    links.append((link_text, full_url))
        
        return links
    
    def extract_city_links(self, html: str, base_url: str) -> List[Dict[str, str]]:
        """Extract city links with optimized parsing"""
        soup = BeautifulSoup(html, 'lxml')
        
        # More specific selector
        city_cubes = soup.select('.city-cube.rounded.white')
        
        links = []
        for cube in city_cubes:
            for link in cube.select('a[href]'):
                link_text = link.get_text(strip=True)
                href = link.get('href')
                
                if not link_text or not href:
                    continue
                
                full_url = urljoin(base_url, href)
                
                # Skip search links
                if '/search' in full_url:
                    continue
                
                links.append({
                    "name": link_text,
                    "link": full_url
                })
        
        return links
    
    def extract_course_details(self, html: str) -> List[Dict[str, str]]:
        """Optimized course detail extraction"""
        soup = BeautifulSoup(html, 'lxml')
        
        # More specific selector
        course_elements = soup.select('.columns.medium-6.large-8.course-details.course-info-wrapper.right-border')
        
        courses = []
        
        for element in course_elements:
            try:
                # Extract course name efficiently
                course_name = ""
                name_elem = element.find(['h1', 'h2', 'h3', 'h4'])
                if name_elem:
                    course_name = name_elem.get_text(strip=True)
                else:
                    link_elem = element.find('a')
                    if link_elem:
                        course_name = link_elem.get_text(strip=True)
                
                if not course_name:
                    continue
                
                # Extract address efficiently
                address = self._extract_address(element)
                
                courses.append({
                    "course_name": course_name,
                    "address": address
                })
                
            except Exception as e:
                print(f"    Error extracting course details: {e}")
                continue
        
        return courses
    
    def _extract_address(self, element) -> str:
        """Optimized address extraction"""
        # Try specific address tag first
        address_elem = element.find('address', {'itemprop': 'address'})
        
        if address_elem:
            return self._clean_address(str(address_elem))
        
        # Fallback to any address tag
        address_elem = element.find('address')
        if address_elem:
            return self._clean_address(str(address_elem))
        
        return "Address not found"
    
    def _clean_address(self, address_html: str) -> str:
        """Clean and format address from HTML"""
        # Replace <br> tags with commas
        cleaned_html = re.sub(r'<br\s*/?>', ', ', address_html, flags=re.IGNORECASE)
        
        # Extract text
        soup = BeautifulSoup(cleaned_html, 'lxml')
        address = soup.get_text(strip=True)
        
        # Clean up formatting
        address = re.sub(r',\s*,', ',', address)  # Remove double commas
        address = re.sub(r'\s*,\s*', ', ', address)  # Normalize comma spacing
        
        return address
    
    def process_city_batch(self, cities_batch: List[Dict]) -> List[Dict]:
        """Process a batch of cities concurrently"""
        all_courses = []
        
        def process_single_city(city_data):
            city_name = city_data.get('name', 'Unknown')
            city_link = city_data.get('link', '')
            destination = city_data.get('destination', 'Unknown')
            
            if not city_link:
                return []
            
            # Rate limiting
            time.sleep(self.delay)
            
            city_html = self.get_html(city_link)
            if not city_html:
                print(f"  Failed to fetch HTML for {city_name}")
                return []
            
            courses = self.extract_course_details(city_html)
            
            # Add metadata to each course
            for course in courses:
                course.update({
                    'city': city_name,
                    'destination': destination,
                    'city_url': city_link
                })
            
            print(f"  {city_name}: Found {len(courses)} courses")
            return courses
        
        # Process cities concurrently
        with concurrent.futures.ThreadPoolExecutor(max_workers=self.max_workers) as executor:
            future_to_city = {
                executor.submit(process_single_city, city_data): city_data 
                for city_data in cities_batch
            }
            
            for future in concurrent.futures.as_completed(future_to_city):
                courses = future.result()
                all_courses.extend(courses)
        
        return all_courses
    
    def save_json(self, data: List[Dict], filename: str):
        """Save data to JSON file efficiently"""
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False, separators=(',', ': '))
    
    def close(self):
        """Clean up resources"""
        self.session.close()

# Legacy functions for compatibility
def get_html(url):
    """Legacy function - creates new scraper instance"""
    scraper = OptimizedGolfScraper()
    result = scraper.get_html(url)
    scraper.close()
    return result

def extract_destination_links(html, base_url):
    """Legacy function"""
    scraper = OptimizedGolfScraper()
    result = scraper.extract_destination_links(html, base_url)
    scraper.close()
    return result

def extract_city_links(html, base_url):
    """Legacy function"""
    scraper = OptimizedGolfScraper()
    result = scraper.extract_city_links(html, base_url)
    scraper.close()
    return result

def extract_course_details(html, base_url):
    """Legacy function"""
    scraper = OptimizedGolfScraper()
    result = scraper.extract_course_details(html)
    scraper.close()
    return result

def process_city_links_file(input_file="all_city_links.json", output_file="course_details.json", test_limit=1, max_workers=3):
    """
    Optimized version of city links processing with concurrent execution
    """
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            city_links = json.load(f)
    except FileNotFoundError:
        print(f"Error: {input_file} not found")
        return 0
    except json.JSONDecodeError:
        print(f"Error: {input_file} is not valid JSON")
        return 0
    
    print(f"Loaded {len(city_links)} city links from {input_file}")
    print(f"Processing {test_limit} cities with {max_workers} concurrent workers")
    
    # Initialize optimized scraper
    scraper = OptimizedGolfScraper(max_workers=max_workers, delay=0.3)
    
    try:
        # Limit cities for testing
        cities_to_process = city_links[:test_limit]
        
        # Process cities in batches
        all_courses = scraper.process_city_batch(cities_to_process)
        
        # Save results
        scraper.save_json(all_courses, output_file)
        
        print(f"\n{'='*60}")
        print(f"SUMMARY: Processed {len(all_courses)} courses from {len(cities_to_process)} cities")
        print(f"Results saved to: {output_file}")
        print(f"{'='*60}")
        
        return len(all_courses)
        
    finally:
        scraper.close()

def process_all_cities_safely(input_file="all_city_links.json", 
                             output_file="all_course_details.json",
                             batch_size=50, 
                             max_workers=3,
                             delay_between_batches=10,
                             checkpoint_interval=500):
    """
    Process all cities in safe batches with checkpointing and progress saving
    
    Args:
        input_file (str): JSON file with city links
        output_file (str): Output file for course details
        batch_size (int): Cities per batch (default: 50)
        max_workers (int): Concurrent workers per batch (default: 3)
        delay_between_batches (int): Seconds to wait between batches (default: 10)
        checkpoint_interval (int): Save progress every N cities (default: 500)
    """
    import os
    import datetime
    
    # Load city links
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            all_city_links = json.load(f)
    except FileNotFoundError:
        print(f"Error: {input_file} not found")
        return 0
    
    total_cities = len(all_city_links)
    print(f"🚀 Starting large-scale processing of {total_cities} cities")
    print(f"📦 Batch size: {batch_size}")
    print(f"🔧 Workers per batch: {max_workers}")
    print(f"⏱️  Delay between batches: {delay_between_batches}s")
    print(f"💾 Checkpoint every: {checkpoint_interval} cities")
    
    # Check for existing progress
    checkpoint_file = f"{output_file}.checkpoint"
    resume_from = 0
    all_courses = []
    
    if os.path.exists(checkpoint_file):
        try:
            with open(checkpoint_file, 'r') as f:
                checkpoint_data = json.load(f)
                resume_from = checkpoint_data.get('last_processed', 0)
                print(f"📍 Resuming from city {resume_from}")
        except:
            print("⚠️  Checkpoint file corrupted, starting fresh")
    
    # Load existing results if resuming
    if resume_from > 0 and os.path.exists(output_file):
        try:
            with open(output_file, 'r') as f:
                all_courses = json.load(f)
                print(f"📂 Loaded {len(all_courses)} existing courses")
        except:
            print("⚠️  Output file corrupted, starting fresh")
            all_courses = []
            resume_from = 0
    
    # Initialize scraper
    scraper = OptimizedGolfScraper(max_workers=max_workers, delay=0.5)
    
    try:
        start_time = datetime.datetime.now()
        processed_cities = resume_from
        
        # Process in batches
        for batch_start in range(resume_from, total_cities, batch_size):
            batch_end = min(batch_start + batch_size, total_cities)
            batch_cities = all_city_links[batch_start:batch_end]
            
            current_batch = (batch_start // batch_size) + 1
            total_batches = (total_cities + batch_size - 1) // batch_size
            
            print(f"\n🔄 Processing batch {current_batch}/{total_batches}")
            print(f"   Cities {batch_start + 1}-{batch_end} of {total_cities}")
            
            # Process this batch
            batch_courses = scraper.process_city_batch(batch_cities)
            all_courses.extend(batch_courses)
            processed_cities = batch_end
            
            # Calculate progress and ETA
            elapsed = datetime.datetime.now() - start_time
            progress_pct = (processed_cities / total_cities) * 100
            
            if processed_cities > resume_from:
                cities_per_second = (processed_cities - resume_from) / elapsed.total_seconds()
                remaining_cities = total_cities - processed_cities
                eta_seconds = remaining_cities / cities_per_second if cities_per_second > 0 else 0
                eta = datetime.timedelta(seconds=int(eta_seconds))
                
                print(f"📊 Progress: {progress_pct:.1f}% ({processed_cities}/{total_cities})")
                print(f"⏰ ETA: {eta} | Speed: {cities_per_second:.2f} cities/sec")
                print(f"🏌️ Total courses found: {len(all_courses)}")
            
            # Save checkpoint
            if processed_cities % checkpoint_interval == 0 or batch_end == total_cities:
                print("💾 Saving checkpoint...")
                
                # Save courses
                with open(output_file, 'w', encoding='utf-8') as f:
                    json.dump(all_courses, f, indent=2, ensure_ascii=False)
                
                # Save checkpoint
                checkpoint_data = {
                    'last_processed': processed_cities,
                    'total_courses': len(all_courses),
                    'timestamp': datetime.datetime.now().isoformat()
                }
                with open(checkpoint_file, 'w') as f:
                    json.dump(checkpoint_data, f, indent=2)
                
                print(f"✅ Saved {len(all_courses)} courses to {output_file}")
            
            # Wait between batches (except for the last one)
            if batch_end < total_cities:
                print(f"😴 Waiting {delay_between_batches}s before next batch...")
                time.sleep(delay_between_batches)
        
        # Final save and cleanup
        total_time = datetime.datetime.now() - start_time
        print(f"\n🎉 COMPLETED!")
        print(f"📊 Processed {total_cities} cities in {total_time}")
        print(f"🏌️ Found {len(all_courses)} total courses")
        print(f"📁 Saved to: {output_file}")
        
        # Clean up checkpoint file
        if os.path.exists(checkpoint_file):
            os.remove(checkpoint_file)
            
        return len(all_courses)
        
    except KeyboardInterrupt:
        print(f"\n⏹️  Process interrupted by user")
        print(f"💾 Saving progress... ({len(all_courses)} courses)")
        
        # Save current progress
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(all_courses, f, indent=2, ensure_ascii=False)
        
        checkpoint_data = {
            'last_processed': processed_cities,
            'total_courses': len(all_courses),
            'timestamp': datetime.datetime.now().isoformat(),
            'interrupted': True
        }
        with open(checkpoint_file, 'w') as f:
            json.dump(checkpoint_data, f, indent=2)
        
        print(f"✅ Progress saved. Resume with same command.")
        return len(all_courses)
        
    except Exception as e:
        print(f"❌ Error occurred: {e}")
        print(f"💾 Saving progress before exit...")
        
        # Save what we have
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(all_courses, f, indent=2, ensure_ascii=False)
        
        return len(all_courses)
        
    finally:
        scraper.close()

def estimate_processing_time(total_cities=9000, batch_size=50, max_workers=3, delay_between_batches=10):
    """
    Estimate how long it will take to process all cities
    """
    import datetime
    
    # Rough estimates based on testing
    cities_per_batch_time = 30  # seconds for 50 cities with 3 workers
    batches_needed = (total_cities + batch_size - 1) // batch_size
    
    processing_time = batches_needed * cities_per_batch_time
    delay_time = (batches_needed - 1) * delay_between_batches
    total_seconds = processing_time + delay_time
    
    total_time = datetime.timedelta(seconds=total_seconds)
    
    print(f"⏱️  PROCESSING TIME ESTIMATE:")
    print(f"   Total cities: {total_cities}")
    print(f"   Batches needed: {batches_needed}")
    print(f"   Estimated time: {total_time}")
    print(f"   (This is a rough estimate - actual time may vary)")

def main():
    """Main function with optimizations"""
    base_url = "https://www.golfnow.com/course-directory/us"
    print(f"Fetching HTML from: {base_url}")
    
    # Initialize scraper
    scraper = OptimizedGolfScraper(delay=1)
    
    try:
        # Get main page
        html = scraper.get_html(base_url)
        if not html:
            print("Failed to fetch main page HTML")
            return
        
        print(f"Successfully fetched {len(html)} characters from main page")
        
        # Extract destination links
        destination_links = scraper.extract_destination_links(html, base_url)
        print(f"Found {len(destination_links)} destination links")
        
        all_city_links = []
        
        # Process destinations with progress
        for i, (dest_name, dest_url) in enumerate(destination_links[:5], 1):  # Limit for demo
            print(f"\n{i}/5 Processing: {dest_name}")
            
            time.sleep(scraper.delay)
            dest_html = scraper.get_html(dest_url)
            
            if dest_html:
                city_links = scraper.extract_city_links(dest_html, dest_url)
                print(f"  Found {len(city_links)} city links")
                
                for city_data in city_links:
                    city_entry = {
                        "destination": dest_name,
                        "name": city_data["name"],
                        "link": city_data["link"]
                    }
                    all_city_links.append(city_entry)
            else:
                print(f"  Failed to fetch HTML for {dest_name}")
        
        # Save results
        scraper.save_json(all_city_links, "all_city_links.json")
        print(f"\nResults saved to: all_city_links.json")
        print(f"Total city links: {len(all_city_links)}")
        
    finally:
        scraper.close()

if __name__ == "__main__":
    import sys
    
    if len(sys.argv) > 1:
        command = sys.argv[1]
        
        if command == "--filter":
            # Small-scale testing
            test_limit = int(sys.argv[2]) if len(sys.argv) > 2 else 1
            max_workers = int(sys.argv[3]) if len(sys.argv) > 3 else 3
            process_city_links_file(test_limit=test_limit, max_workers=max_workers)
            
        elif command == "--all":
            # Large-scale processing of all cities
            batch_size = int(sys.argv[2]) if len(sys.argv) > 2 else 50
            max_workers = int(sys.argv[3]) if len(sys.argv) > 3 else 3
            delay_between_batches = int(sys.argv[4]) if len(sys.argv) > 4 else 10
            
            print("🚨 LARGE-SCALE PROCESSING MODE")
            print("This will process ALL cities in your JSON file.")
            print("Make sure you have a stable internet connection!")
            print("\nPress Ctrl+C anytime to safely stop and save progress.")
            
            response = input("\nContinue? (y/N): ").strip().lower()
            if response == 'y':
                process_all_cities_safely(
                    batch_size=batch_size, 
                    max_workers=max_workers,
                    delay_between_batches=delay_between_batches
                )
            else:
                print("❌ Cancelled")
                
        elif command == "--estimate":
            # Estimate processing time
            total_cities = int(sys.argv[2]) if len(sys.argv) > 2 else 9000
            batch_size = int(sys.argv[3]) if len(sys.argv) > 3 else 50
            max_workers = int(sys.argv[4]) if len(sys.argv) > 4 else 3
            delay_between_batches = int(sys.argv[5]) if len(sys.argv) > 5 else 10
            
            estimate_processing_time(total_cities, batch_size, max_workers, delay_between_batches)
            
        else:
            print("Unknown command. Available commands:")
            print("  --filter [limit] [workers]     : Test with limited cities")
            print("  --all [batch_size] [workers] [delay] : Process all cities")
            print("  --estimate [total] [batch] [workers] [delay] : Estimate time")
    else:
        main()
