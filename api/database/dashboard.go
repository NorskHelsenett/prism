package database

import (
	"encoding/json"
)

func CountJSONData(year string) (int64, error) {
	var count int64
	query := db.Model(&JSONData{})
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	result := query.Count(&count)
	return count, result.Error
}

func CountProjects() (int64, error) {
	var count int64
	result := db.Model(&ProjectData{}).Count(&count)
	return count, result.Error
}

func CountBugBounties(year string) (int64, error) {
	var count int64
	var bugBountyProjectIDs []uint

	// Fetch IDs of projects with bug bounties
	err := db.Model(&ProjectData{}).Where("is_bug_bounty = ?", true).Pluck("id", &bugBountyProjectIDs).Error
	if err != nil {
		return 0, err
	}

	// Count vulnerabilities associated with those projects
	query := db.Model(&JSONData{}).Where("project_id IN (?)", bugBountyProjectIDs)
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	err = query.Count(&count).Error
	return count, err
}

func CountCriticalities(year string) (map[string]int, error) {
	var jsonData []JSONData
	query := db
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	result := query.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	criticalityCounts := make(map[string]int)
	for _, data := range jsonData {
		var vuln Vulnerability
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			continue
		}
		criticalityCounts[vuln.Criticality]++
	}
	return criticalityCounts, nil
}

func CountOWASPCategories(year string) (map[string]int, error) {
	var jsonData []JSONData
	query := db
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	result := query.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	categoryCounts := make(map[string]int)
	for _, data := range jsonData {
		var vuln Vulnerability
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			continue
		}
		category := vuln.Category
		if category == "" {
			category = "uncategorized"
		}
		categoryCounts[category]++
	}
	return categoryCounts, nil
}

func FetchOWASPCriticalities(year string) (map[string]map[string]int, error) {
	var jsonData []JSONData
	query := db
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	result := query.Find(&jsonData)
	if result.Error != nil {
		return nil, result.Error
	}

	owaspData := make(map[string]map[string]int)
	for _, data := range jsonData {
		var vuln VulnerabilityData
		err := json.Unmarshal(data.Vulnerability, &vuln)
		if err != nil {
			return nil, err
		}

		category := vuln.Category
		if category == "" {
			category = "Uncategorized"
		}

		if _, exists := owaspData[category]; !exists {
			owaspData[category] = map[string]int{
				"information": 0,
				"low":        0,
				"medium":     0,
				"high":       0,
				"critical":   0,
			}
		}

		switch vuln.Criticality {
		case "information", "low", "medium", "high", "critical":
			owaspData[category][vuln.Criticality]++
		}
	}
	return owaspData, nil
}

func CountByStatus(year string) (map[string]int, error) {
	var results []struct {
		Status string
		Count  int
	}
	query := db.Model(&JSONData{})
	if year != "" {
		query = query.Where("strftime('%Y', created_at) = ?", year)
	}
	
	if err := query.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	statusCounts := make(map[string]int)
	for _, result := range results {
		statusCounts[result.Status] = result.Count
	}
	return statusCounts, nil
}
