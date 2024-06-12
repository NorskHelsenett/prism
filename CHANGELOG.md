# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.6] - 2024-06-12

### Added

- Toast 🥪

  We have a new toast-master! Using the amazing [svelte-sonner](https://github.com/wobsoriano/svelte-sonner) we now have an awesome toaster information experience like no other!

### Fixes

- With our latest update to the `/all` vulnerabilities endpoint, we've significantly reduced the data footprint of the responses. This change is part of our ongoing commitment to environmental sustainability and digital efficiency. This change is a step forward in our journey to support global climate health while continuing to provide our users with the high-quality service they expect.

## [0.1.5] - 2024-06-10

### Added

- Profile picture 👨‍🦳

  Profile picture is not downloaded from Azure GraphQL endpoint and populated. No more missing and empty images.

- Delete User ❌

  It is now possible to delete a user from the system. This feature allows for the removal of unnecessary or inactive users, maintaining the integrity and security of the system.

## [0.1.4] - 2024-06-10

### Added

- NOTIFICATIONS 📣

  Notifications is now implemented with the support of browser push notifications. A list of your recent notifications is easily accessible from the top right corner of the application. Indulge in a serene experience and acknowledge each new notification as it arrives, today!

## [0.1.3] - 2024-04-13

### Added

- Planning is now implemented

## [0.1.3] - 2024-02-13

### Added

- Comments are now implemented with edit, reply to and delete.
- Activites, a historic view of the activities are shown

## [0.1.2] - 2024-02-13

### Added

- Field for issue tracker url in vulnerabilities
- Copy title and text formatted for external issue systems
- Copy one image. Only one image is allowed in the Clipboard API for browsers. Firefox is not supported, only Chromium.
- Image carousel for images in vulnerability
- Copy markdown image format when clicking on images

### Fixed

- Images are now stored at maximum 2160px - up from 1280px.

## [0.1.1] - 2024-01-30

### Added

- RBAC and ReBAC added. Roles are managed with `roles.yaml` and each access is fine-tuned to each resource and action (`read`, `write`, `delete`) which translates to `GET`, `POST/PUT/PATCH`, `DELETE` HTTP methods.
- User View in Settings, now you can assign roles in the settings if you have RBAC access
- OTP is now a switch in the settings panel, turn it on-off again.
- [CI] Pipeline produces helm artifacts for PROD and feature branches respectively

### Fixed

- Bug in images are fixed, now edited vulnerabilites does not delete images

### Changed

- Config is now loaded into memory at init

## [0.1.1] - 2024-01-26

### Added

- Todo support for Project Description and in vulnerability

## [0.1.0] - 2024-01-19 - Morris Release

### Added

- OTP (One-Time Password) implementation for enhanced user security.
- Server-side session management for robust session control.
- Database optimization for improved performance.
- Comprehensive audit logging for all requests.
- Slack notifications for events triggered by new findings.
- Event-driven mechanism with SQLite database table monitoring.
- Display of OWASP Categories by Severity on the dashboard.
- Spidergraph visualization for all vulnerability statuses.
- Markdown support in TextArea inputs for Projects and Vulnerabilities.

### Security

- Completion of penetration testing, leading to the identification and resolution of security vulnerabilities.
- Addressed several vulnerabilities to strengthen system integrity and user data protection.

## [0.0.4] - 2024-01-10

### Added

- Unified vulnerability table
- SORTING!: Now sort your vulnerability table (dummy way)
- Added OWASP Category top 3 Doughnut Chart
- About page with some curious info
- Vulnerability status
- Edit your projects
- Edit your vulnerability findings (not for images yet)

### Removed

- Icon from Create Project

## [0.0.3] - 2024-01-09

### Added

- Usersearch component to search and add multiple users to projects or other use-cases.
- HostRule to restrict access to our fqdn by IP whitelist

### Fixed

- Vulnerability views are now sorted by newest first
- Importing not working as expected

## [0.0.2] - 2024-01-08

### Added

- Export/Import functionality
- Settings panel
- Profile settings panel (view-only)

## [0.0.1] - 2023-12-06

### Added

- Initial release of the project.
- Home Dashboard
- Vulnerabilites Dashboard
- Projects Dashboard
