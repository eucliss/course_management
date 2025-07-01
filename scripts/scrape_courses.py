#!/usr/bin/env python3
"""
Script to get HTML and extract links from us-destination-wrapper class
"""

import requests
from bs4 import BeautifulSoup
from urllib.parse import urljoin

def get_html(url):
    """
    Get HTML content from a URL
    
    Args:
        url (str): The URL to fetch
        
    Returns:
        str: HTML content or None if failed
    """
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36'
    }
    
    try:
        response = requests.get(url, headers=headers, timeout=10)
        response.raise_for_status()
        return response.text
    except requests.RequestException as e:
        print(f"Error fetching {url}: {e}")
        return None

def extract_destination_links(html, base_url):
    """
    Extract all links from elements with class 'us-destination-wrapper white rounded'
    
    Args:
        html (str): HTML content
        base_url (str): Base URL for resolving relative links
        
    Returns:
        list: List of tuples (link_text, full_url)
    """
    soup = BeautifulSoup(html, 'html.parser')
    
    # Find elements with the specific class
    destination_wrappers = soup.find_all(class_="us-destination-wrapper white rounded")
    
    links = []
    
    for wrapper in destination_wrappers:
        # Find all links within this wrapper
        for link in wrapper.find_all('a', href=True):
            link_text = link.get_text(strip=True)
            href = link.get('href')
            
            # Convert relative URLs to absolute URLs
            full_url = urljoin(base_url, href)
            
            if link_text and href:
                links.append((link_text, full_url))
    
    return links

def main():
    url = "https://www.golfnow.com/course-directory/us"
    print(f"Fetching HTML from: {url}")
    
    html = get_html(url)
    
    if html:
        print(f"Successfully fetched {len(html)} characters")
        
        # Extract links from us-destination-wrapper elements
        links = extract_destination_links(html, url)
        
        print(f"\nFound {len(links)} links in 'us-destination-wrapper white rounded' elements:")
        print("-" * 60)
        
        for i, (text, link_url) in enumerate(links, 1):
            print(f"{i:2d}. {text}")
            print(f"    {link_url}")
            print()
        
        # Save links to file
        with open("destination_links.txt", "w", encoding="utf-8") as f:
            for text, link_url in links:
                f.write(f"{text}\n{link_url}\n\n")
        
        print(f"Links saved to: destination_links.txt")
        
    else:
        print("Failed to fetch HTML")

if __name__ == "__main__":
    main()
