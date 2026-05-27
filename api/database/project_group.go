package database

import "gorm.io/gorm"

func ListProjectGroups() ([]ProjectGroup, error) {
	var groups []ProjectGroup
	result := db.Order("sort_order ASC, id ASC").Find(&groups)
	return groups, result.Error
}

func CreateProjectGroup(group *ProjectGroup) error {
	return db.Create(group).Error
}

func UpdateProjectGroup(group *ProjectGroup) error {
	return db.Model(&ProjectGroup{}).Where("id = ?", group.ID).
		Updates(map[string]interface{}{
			"name":       group.Name,
			"color":      group.Color,
			"sort_order": group.SortOrder,
		}).Error
}

func DeleteProjectGroup(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Move members back to ungrouped
		if err := tx.Model(&ProjectData{}).
			Where("group_id = ?", id).
			Update("group_id", nil).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", id).Delete(&ProjectGroup{}).Error
	})
}

func SetProjectGroup(projectID uint, groupID *uint, sortOrder *int) error {
	// Verify the project exists (and isn't soft-deleted) before issuing the
	// update so the caller can distinguish "not found" from "DB error".
	var count int64
	if err := db.Model(&ProjectData{}).Where("id = ?", projectID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	updates := map[string]interface{}{
		"group_id": groupID,
	}
	if sortOrder != nil {
		updates["sort_order"] = *sortOrder
	}
	return db.Model(&ProjectData{}).Where("id = ?", projectID).Updates(updates).Error
}
