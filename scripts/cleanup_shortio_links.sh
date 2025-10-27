#!/bin/bash

# Script to remove all existing links for morph-service.short.gy domain using Short.io API
# Requires MORPH_REDIRECT_KEY environment variable
#
# Usage:
#   export MORPH_REDIRECT_KEY="your_api_key_here"
#   ./scripts/cleanup_shortio_links.sh

# set -e  # Commented out to prevent silent exits

# Configuration
DOMAIN="morph-service.short.gy"
API_BASE="https://api.short.io/api"
DELETE_API_BASE="https://api.short.io"
LIMIT=100

# Check if required environment variable is set
if [ -z "$MORPH_REDIRECT_KEY" ]; then
    echo "Error: MORPH_REDIRECT_KEY environment variable is not set"
    echo "Please set it with: export MORPH_REDIRECT_KEY='your_api_key_here'"
    exit 1
fi

# Function to get domain ID from domain name
get_domain_id() {
    local domain_name="$1"
    echo "Getting domain ID for: $domain_name" >&2

    local response=$(curl -s -H "Authorization: $MORPH_REDIRECT_KEY" \
        "$API_BASE/domains")


    # Check if response is valid JSON
    if ! echo "$response" | jq empty 2>/dev/null; then
        echo "Error: Invalid JSON response from domains API" >&2
        echo "Response: $response" >&2
        echo "This might be an authentication issue. Please check your MORPH_REDIRECT_KEY." >&2
        return 1
    fi

    # Extract domain ID
    local domain_id=$(echo "$response" | jq -r --arg domain "$domain_name" \
        '.[] | select(.hostname == $domain) | .id')

    if [ "$domain_id" = "null" ] || [ -z "$domain_id" ]; then
        echo "Error: Could not find domain ID for $domain_name" >&2
        echo "Available domains:" >&2
        echo "$response" | jq -r '.[].hostname // "none"' >&2
        return 1
    fi

    echo "Found domain ID: $domain_id" >&2
    echo "$domain_id"
}

# Function to get links for a domain
get_links() {
    local domain_id="$1"
    local limit="$2"

    echo "Fetching up to $limit links for domain ID: $domain_id" >&2

    local response=$(curl -s -H "Authorization: $MORPH_REDIRECT_KEY" \
        "$API_BASE/links?domain_id=$domain_id&limit=$limit")


    # Check if response is valid JSON
    if ! echo "$response" | jq empty 2>/dev/null; then
        echo "Error: Invalid JSON response from links API" >&2
        echo "Response: $response" >&2
        return 1
    fi

    # Check if request was successful
    local status_code=$(echo "$response" | jq -r '.statusCode // 200')
    if [ "$status_code" != "200" ]; then
        echo "Error: Failed to fetch links. Status: $status_code" >&2
        echo "Response: $response" >&2
        return 1
    fi

    echo "$response"
}

# Function to delete links one by one
delete_links() {
    local link_ids="$1"

    if [ -z "$link_ids" ]; then
        echo "No link IDs to delete"
        return 0
    fi

    echo "Deleting links: $link_ids"

    # Split comma-separated string into array
    IFS=',' read -ra LINK_ARRAY <<< "$link_ids"

    local deleted_count=0
    local failed_count=0

    for link_id in "${LINK_ARRAY[@]}"; do
        if [ -n "$link_id" ]; then
            echo "Deleting link: $link_id"
            local response=$(curl -s -X DELETE \
                -H "Authorization: $MORPH_REDIRECT_KEY" \
                "$DELETE_API_BASE/links/$link_id")


            # Check if request was successful
            local success=$(echo "$response" | jq -r '.success // false')
            if [ "$success" = "true" ]; then
                echo "Successfully deleted: $link_id"
                deleted_count=$((deleted_count + 1))
            else
                echo "Failed to delete: $link_id"
                echo "Response: $response"
                failed_count=$((failed_count + 1))
            fi

            # Small delay between deletions to be respectful to the API
            sleep 0.1
        fi
    done

    echo "Deletion summary: $deleted_count successful, $failed_count failed"

    if [ $failed_count -gt 0 ]; then
        return 1
    fi
}

# Main cleanup process
main() {
    echo "Starting cleanup of Short.io links for domain: $DOMAIN"
    echo "================================================"

    # Get domain ID
    echo "Step 1: Getting domain ID..."
    DOMAIN_ID=$(get_domain_id "$DOMAIN")
    if [ $? -ne 0 ] || [ -z "$DOMAIN_ID" ]; then
        echo "Failed to get domain ID. Exiting."
        exit 1
    fi
    echo "Using domain ID: $DOMAIN_ID"

    # Initialize counters
    total_deleted=0
    batch_count=0

    # Main cleanup loop
    while true; do
        batch_count=$((batch_count + 1))
        echo ""
        echo "--- Batch $batch_count ---"

        # Get links for this batch
        echo "Calling get_links with domain_id: $DOMAIN_ID, limit: $LIMIT"
        links_response=$(get_links "$DOMAIN_ID" "$LIMIT")
        echo "Links response received, length: ${#links_response}"

        # Extract link IDs from response
        # First check if links array exists and is not null
        local links_exists=$(echo "$links_response" | jq -r '.links // null')
        if [ "$links_exists" = "null" ]; then
            echo "No links array found in response. Cleanup complete!"
            break
        fi

        # Check if links array is empty
        local links_count=$(echo "$links_response" | jq -r '.links | length' 2>/dev/null || echo "0")
        if [ "$links_count" = "0" ] || [ "$links_count" = "null" ]; then
            echo "No more links found. Cleanup complete!"
            break
        fi

        link_ids=$(echo "$links_response" | jq -r '.links[].id' 2>/dev/null | tr '\n' ',' | sed 's/,$//')

        # Check if we have any links to delete
        if [ -z "$link_ids" ]; then
            echo "No more links found. Cleanup complete!"
            break
        fi

        # Count links in this batch
        link_count=$links_count
        echo "Found $link_count links in this batch"

        # Delete the links
        delete_links "$link_ids"

        # Update total count
        total_deleted=$((total_deleted + link_count))
        echo "Total links deleted so far: $total_deleted"

        # Small delay to be respectful to the API
        sleep 1
    done

    echo ""
    echo "================================================"
    echo "Cleanup completed successfully!"
    echo "Total links deleted: $total_deleted"
    echo "Total batches processed: $batch_count"
}

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed."
    echo "Please install jq: brew install jq (on macOS) or apt-get install jq (on Ubuntu)"
    exit 1
fi

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo "Error: curl is required but not installed."
    exit 1
fi

# Run the main function
main