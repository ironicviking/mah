# Changelog

All notable changes to MAH (Multi-Architecture Hub) will be documented in this file.

## [2.0.3] - 2025-09-03

### Fixed
- Fixed Docker GPG key installation permissions issue on Ubuntu and Debian
- Changed sudo redirection approach to use `sudo tee` instead of direct redirection
- Resolves "Permission denied" error when writing to `/usr/share/keyrings/`

## [2.0.2] - Previous Release
- Previous stable release baseline