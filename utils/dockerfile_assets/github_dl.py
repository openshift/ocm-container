#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import argparse
import time
import requests
import hashlib

# Install of the various components for ocm-container/backplane-tools requires
# about 50 available quota to complete. Adjust this number if adding additional
# tooling pushes past this value in the future
MINIMUM_QUOTA_REQUIREMENT = 50


def validate_binary(binary, checksum, raw_algorithm="sha256") -> bool:
    hash_function = None
    expected_hash, _ = checksum.split()

    algorithm = raw_algorithm.removesuffix("sum").lower()

    if algorithm == "sha256":
        hash_function = hashlib.sha256
    elif algorithm == "md5":
        hash_function = hashlib.md5
    else:
        print(f"Unsupported hash algorithm: {raw_algorithm}")
        return False

    # Calculate the hash of the binary
    hash_object = hash_function(binary)
    calculated_hash = hash_object.hexdigest()

    if calculated_hash != expected_hash:
        print(f"Checksum validation failed: expected {expected_hash}, got {calculated_hash}")
        return False

    print("Checksum validation succeeded.")
    return True


def get_url_with_authentication(url, token=None, additional_headers=None, retry=0, max_retries=5) -> requests.Response | None:
    if retry > max_retries:
        print("max retries reached. Exiting")
        return None

    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    if additional_headers:
        headers.update(additional_headers)

    try:
        response = requests.get(url, headers=headers, timeout=30)
    except requests.exceptions.RequestException as e:
        print(f"Request failed for {url}: {e}")
        return None

    if response.status_code == 200:
        return response

    if response.status_code >= 500:
        print(f"Got {response.status_code}. Backing-off and retrying...")
        retry += 1
        time.sleep(3 * retry)

        return get_url_with_authentication(url, token, additional_headers, retry, max_retries)

    print(f"Failed to fetch data from {url}: {response.status_code} {response.text}")
    return None


def validate_token(token) -> bool:
    additional_headers = {
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }

    response = get_url_with_authentication(
        "https://api.github.com/rate_limit", token, additional_headers,
    )

    if response is None:
        print("Error: Failed to validate GitHub token (no response from API)")
        return False

    try:
        remaining = response.json().get("rate", {}).get("remaining", 0)
    except (ValueError, AttributeError):
        print("Error: Malformed response from GitHub rate_limit API")
        return False

    if not isinstance(remaining, int):
        print("Error: Malformed response from GitHub rate_limit API")
        return False

    if remaining < MINIMUM_QUOTA_REQUIREMENT:
        print(f"Error: GitHub API rate limit nearly exhausted ({remaining} remaining, need at least {MINIMUM_QUOTA_REQUIREMENT} for a full build)")
        return False

    print(f"GitHub token authenticated successfully (API calls remaining: {remaining})")
    return True


def list_assets(url, token=None) -> list:
    response = get_url_with_authentication(url, token)
    if response is None:
        print(f"Failed to fetch content from {url}")
        return []

    content = response.json()
    if not content:
        print(f"Failed to fetch content from {url}")
        return []
    
    if "assets" not in content:
        print(f"No assets found in the release at {url}")
        return []

    return content.get("assets", [])


def extract_browser_download_url(assets, asset) -> str:
    for item in assets:
        if item.get("name") == asset:
            return item.get("browser_download_url")

    print(f"Asset '{asset}' not found in the release")
    print("Available assets:")
    for item in assets:
        print(f"\t{item.get('name')}")

    return ""


def get_checksum(assets, checksum_file, platform, token=None) -> str:
    checksum_download_url = extract_browser_download_url(assets, checksum_file)
    if not checksum_download_url:
        print(f"{checksum_file} not found")
        return ""

    print(f"Downloading checksum file from {checksum_download_url}")
    response = get_url_with_authentication(checksum_download_url, token)
    if response is None:
        print(f"Failed to download checksum file from {checksum_download_url}")
        return ""
    if not response.content:
        print(f"No content found in {checksum_file}")
        return ""

    checksum_file_content = response.content.decode('utf-8')
    checksum = list(filter(lambda line: platform in line, checksum_file_content.splitlines()))

    if not checksum:
        print(f"No checksum found for platform '{platform}' in {checksum_file}")
        return ""

    if len(checksum) > 1:
        print(f"Multiple checksums found for platform '{platform}' in {checksum_file}:")
        for item in checksum:
            print(f"\t{item}")
        return ""

    return checksum[0].strip()


def get_binary(assets, checksum, token=None) -> bytes:
    binary_name = checksum.split()[1]
    binary_download_url = extract_browser_download_url(assets, binary_name)
    if not binary_download_url:
        print(f"{binary_name} not found")
        return b""

    print(f"Downloading binary from {binary_download_url}")
    response = get_url_with_authentication(binary_download_url, token)
    if response is None:
        print(f"Failed to download binary from {binary_download_url}")
        return b""
    
    if not response.content:
        print(f"No content found for {binary_name}")
        return b""

    return response.content


def get_quota(token=None) -> list | None:
    quota_errors = []

    additional_headers = {
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28"
    }

    response = get_url_with_authentication("https://api.github.com/rate_limit", token, additional_headers)

    if response is None:
        print("Failed to fetch GitHub API rate limit information.")
        return

    print("GitHub API Rate Limit Information:")

    for key, value in response.json()['resources'].items():
        print(f"\n{key.capitalize()} Rate Limit:")

        if value['limit'] == 0:
            quota_errors.append((key, f"total: {value['limit']}, remaining: {value['remaining']}, reset: {value['reset']}"))
        
        for k, v in value.items():
            print(f"{k.capitalize()}: {v}")

    return quota_errors if quota_errors else None


def main():
    parser = argparse.ArgumentParser(
        description="GitHub Downloader",
        epilog="If an (optional) GITHUB_TOKEN environment variable is set, it will be used as a bearer token to authenticate the request.")
    subparsers = parser.add_subparsers(dest="command", required=True)
    subparsers.add_parser("quota", help="Get GitHub API rate limit information")

    download_parser = subparsers.add_parser("download", help="Download a GitHub asset")
    download_parser.add_argument("--url", required=True, help="GitHub asset URL to download (e.g., 'https://api.github.com/repos/ORGANIZATION/REPOSITORY/releases/tags/v1.0.0')")
    download_parser.add_argument("--checksum_file", required=True, help="Name of the checksum file to download (e.g., 'checksums.txt')")
    download_parser.add_argument("--checksum_algorithm", default="sha256", help="Checksum algorithm to use (default: 'sha256') ")
    download_parser.add_argument("--platform", required=True, help="Platform to download assets for (e.g., 'amd64')")
    args = parser.parse_args()
    args.token = None

    # env var mounted as build secret
    secretMount = "/run/secrets/GITHUB_TOKEN"
    if os.path.isfile(secretMount):
        with open(secretMount) as f:
            token = f.read().strip()
            if token:
                args.token = token

    # CI secret
    tokenMount = "/run/secrets/read-only-github-pat/token"
    if os.path.isfile(tokenMount):
        with open(tokenMount) as f:
            token = f.read().strip()
            if token:
                args.token = token

    # env var
    token = os.environ.get("GITHUB_TOKEN", "").strip()
    if token:
        args.token = token

    if args.token is None:
        print("No GITHUB_TOKEN found in environment variables nor secret. Proceeding without authentication.")
    elif not validate_token(args.token):
        sys.exit(1)

    if args.command == "quota":
        errors = get_quota(args.token)
        if errors is not None:
            for error in errors:
                print(f"Quota error: {error[0]} - {error[1]}")
            sys.exit(1)
        
        sys.exit(0)
    
    assets = list_assets(args.url, args.token)
    if not assets:
        sys.exit(1)

    checksum = get_checksum(assets, args.checksum_file, args.platform, args.token)
    if not checksum:
        sys.exit(1)

    binary = get_binary(assets, checksum, args.token)
    if not binary:
        sys.exit(1)

    if not validate_binary(binary, checksum, args.checksum_algorithm):
        sys.exit(1)

    with open(checksum.split()[1], "wb") as f:
        f.write(binary)

    print(f"Binary '{checksum.split()[1]}' downloaded and validated successfully.")
    

if __name__ == "__main__":
    main()
