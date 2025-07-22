#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import argparse
import requests
import hashlib


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


def get_url_with_authentication(url, token=None, additional_headers=None) -> requests.Response:
    headers = {}
    if token:
        headers["Authorization"] = f"Bearer {token}"

    if additional_headers:
        headers.update(additional_headers)
    
    response = requests.get(url, headers=headers)
    
    if response.status_code != 200:
        print(f"Failed to fetch data from {url}: {response.status_code} {response.text}")
        return None

    return response


def list_assets(url, token=None) -> list:
    content = get_url_with_authentication(url, token).json()
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

    # Else, if the asset is not found, print an error message and exit
    print(f"Asset '{asset}' not found in the release")
    print(f"Available assets:")

    [print(f"\t{item.get('name')}") for item in assets]

    return ""


def get_checksum(assets, checksum_file, platform, token=None) -> str:
    checksum = None
    checksum_download_url = extract_browser_download_url(assets, checksum_file)
    if not checksum_download_url:
        print(f"{checksum_file} not found")
        return ""


    print(f"Downloading checksum file from {checksum_download_url}")
    response = get_url_with_authentication(checksum_download_url, token)
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


def get_quota(token=None) -> list:
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

    if os.environ.get("GITHUB_TOKEN"):
        args.token = os.environ["GITHUB_TOKEN"]
    else:
        args.token = None
        print("No GITHUB_TOKEN found in environment variables. Proceeding without authentication.")

    if args.command == "quota":
        errors = get_quota(getattr(args, 'token', None))
        if errors is not None:
            for error in errors:
                print(f"Quota error: {error[0]} - {error[1]}")
            sys.exit(1)
        
        sys.exit(0)
    
    assets = list_assets(args.url, getattr(args, 'token', None))
    if not assets:
        sys.exit(1)

    checksum = get_checksum(assets, args.checksum_file, args.platform, getattr(args, 'token', None))
    if not checksum:
        sys.exit(1)

    binary = get_binary(assets, checksum, getattr(args, 'token', None))
    if not binary:
        sys.exit(1)

    if not validate_binary(binary, checksum, args.checksum_algorithm):
        sys.exit(1)

    with open(checksum.split()[1], "wb") as f:
        f.write(binary)

    print(f"Binary '{checksum.split()[1]}' downloaded and validated successfully.")
    

if __name__ == "__main__":
    main()
