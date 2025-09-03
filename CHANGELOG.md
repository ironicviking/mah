# Changelog

All notable changes to MAH (Multi-Architecture Hub) will be documented in this file.

## [2.0.4] - 2025-09-03

### Fixed
- Fixed comprehensive sudo permission issues in system operations
- Fixed Docker repository setup using sudo tee instead of direct redirection
- Fixed Rocky Linux dnf-automatic config file creation permissions  
- Fixed Docker provider compose and env file creation permissions
- Replaced all 'cat > file' redirections with 'cat | sudo tee file' pattern
- Ensures proper permissions for writing to system directories across all platforms

## [2.0.3] - 2025-09-03

### Fixed
- Fixed Docker GPG key installation permissions issue on Ubuntu and Debian
- Changed sudo redirection approach to use `sudo tee` instead of direct redirection
- Resolves "Permission denied" error when writing to `/usr/share/keyrings/`

## [2.0.2] - Previous Release
- Previous stable release baseline