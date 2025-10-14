package database

import (
	"encoding/json"
	"errors"
)

// DashboardMetrics captures the aggregated data returned by the dashboard endpoint.
type DashboardMetrics struct {
	Total              int64
	Projects           int64
	BugBounties        int64
	Statuses           map[string]int
	Criticalities      map[string]int
	OWASP              map[string]int
	OWASPCriticalities map[string]map[string]int
}

// GetDashboardMetrics returns aggregated dashboard data filtered by what the user is allowed to see.
func GetDashboardMetrics(year, email string, isGlobalVulnerability, isGlobalProject bool) (DashboardMetrics, error) {
	metrics := DashboardMetrics{
		Statuses:           make(map[string]int),
		Criticalities:      make(map[string]int),
		OWASP:              make(map[string]int),
		OWASPCriticalities: make(map[string]map[string]int),
	}

	projects, err := accessibleProjects(email, isGlobalProject)
	if err != nil {
		return metrics, err
	}

	projectIDs := make([]uint, 0, len(projects))
	bugBountyProjects := make(map[uint]bool, len(projects))
	for _, project := range projects {
		projectIDs = append(projectIDs, project.ID)
		if project.IsBugBounty {
			bugBountyProjects[project.ID] = true
		}
	}
	metrics.Projects = int64(len(projects))

	vulnerabilities, err := accessibleVulnerabilities(year, email, isGlobalVulnerability, projectIDs)
	if err != nil {
		return metrics, err
	}

	metrics.Total = int64(len(vulnerabilities))

	for _, entry := range vulnerabilities {
		if entry.Project != nil && entry.Project.IsBugBounty {
			metrics.BugBounties++
		} else if entry.ProjectID != nil && bugBountyProjects[*entry.ProjectID] {
			metrics.BugBounties++
		}

		metrics.Statuses[entry.Status]++

		var vuln Vulnerability
		if err := json.Unmarshal(entry.Vulnerability, &vuln); err == nil {
			criticality := vuln.Criticality
			if criticality == "" {
				criticality = "unknown"
			}
			metrics.Criticalities[criticality]++

			category := vuln.Category
			if category == "" {
				category = "uncategorized"
			}
			metrics.OWASP[category]++

			counts := metrics.OWASPCriticalities[category]
			if counts == nil {
				counts = map[string]int{
					"information": 0,
					"low":         0,
					"medium":      0,
					"high":        0,
					"critical":    0,
				}
			}

			switch vuln.Criticality {
			case "information", "low", "medium", "high", "critical":
				counts[vuln.Criticality]++
			}

			metrics.OWASPCriticalities[category] = counts
		}
	}

	return metrics, nil
}

func accessibleProjects(email string, isGlobal bool) ([]ProjectData, error) {
	switch {
	case isGlobal:
		return GetProjects()
	case email == "":
		return nil, errors.New("email required for project scoped dashboard access")
	default:
		return GetProjectsFor(email)
	}
}

func accessibleVulnerabilities(year, email string, isGlobal bool, projectIDs []uint) ([]JSONData, error) {
	query := db.Preload("Project").Model(&JSONData{})
	if year != "" {
		query = query.Where("strftime('%Y', json_data.created_at) = ?", year)
	}

	var data []JSONData

	if isGlobal {
		if err := query.Find(&data).Error; err != nil {
			return nil, err
		}
		return data, nil
	}

	if email == "" {
		return []JSONData{}, nil
	}

	emailPattern := "%" + email + "%"
	constraints := db.Where("found_by LIKE ?", emailPattern)
	if len(projectIDs) > 0 {
		constraints = constraints.Or("project_id IN ?", projectIDs)
	}

	if err := query.Where(constraints).Find(&data).Error; err != nil {
		return nil, err
	}

	filtered := FilterJSONDataForUser(data, email, false)
	return filtered, nil
}
