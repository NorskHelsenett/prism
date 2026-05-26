package database

import (
	"errors"
	"prism/models"

	"gorm.io/gorm"
)

// CreateReport persists a new Report draft and returns the created row.
// ShareToken is set by the caller (so the route handler can guarantee uniqueness).
func CreateReport(r *models.Report) error {
	return db.Create(r).Error
}

// GetReport loads a report by ID. Returns ErrNotFound (wrapped) if missing.
func GetReport(id uint) (*models.Report, error) {
	var r models.Report
	if err := db.First(&r, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

// GetReportByToken resolves a short token to a Report. Used by the public
// share endpoint.
func GetReportByToken(token string) (*models.Report, error) {
	var r models.Report
	if err := db.Where("share_token = ?", token).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

// ListReports returns every report; ACL filtering is the caller's job
// (HasReadOnAnyProject on each row's ProjectIDsList).
func ListReports() ([]models.Report, error) {
	var rs []models.Report
	if err := db.Order("created_at DESC").Find(&rs).Error; err != nil {
		return nil, err
	}
	return rs, nil
}

// UpdateReport saves the row (BeforeSave hook re-marshals JSON columns).
func UpdateReport(r *models.Report) error {
	return db.Save(r).Error
}

// DeleteReport soft-deletes the draft and hard-deletes its versions.
// Versions are removed because their PDF bytes and frozen JSON have no
// meaning without the parent.
func DeleteReport(id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("report_id = ?", id).Delete(&models.ReportVersion{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Report{}, id).Error
	})
}

// TokenExists checks for a collision before assigning a new share token.
func TokenExists(token string) (bool, error) {
	var count int64
	if err := db.Model(&models.Report{}).Where("share_token = ?", token).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListReportVersions returns every published version of a report, newest first.
func ListReportVersions(reportID uint) ([]models.ReportVersion, error) {
	var vs []models.ReportVersion
	if err := db.Where("report_id = ?", reportID).
		Order("version_number DESC").
		Find(&vs).Error; err != nil {
		return nil, err
	}
	return vs, nil
}

// GetReportVersion returns one version by report id + version number.
func GetReportVersion(reportID uint, version int) (*models.ReportVersion, error) {
	var v models.ReportVersion
	if err := db.Where("report_id = ? AND version_number = ?", reportID, version).First(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

// GetReportVersionByID returns one version by its primary key (used when
// resolving Report.LatestPublishedVersionID).
func GetReportVersionByID(id uint) (*models.ReportVersion, error) {
	var v models.ReportVersion
	if err := db.First(&v, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

// PublishReportVersion runs the publish transaction: it computes the next
// version number, inserts the snapshot, and sets the parent report's
// LatestPublishedVersionID. All in one tx so concurrent publishes can't
// duplicate a version number or leave a stale latest-pointer.
func PublishReportVersion(reportID uint, build func(version int) (*models.ReportVersion, error)) (*models.ReportVersion, error) {
	var saved *models.ReportVersion
	err := db.Transaction(func(tx *gorm.DB) error {
		var last models.ReportVersion
		next := 1
		err := tx.Where("report_id = ?", reportID).
			Order("version_number DESC").
			Limit(1).
			First(&last).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			next = last.VersionNumber + 1
		}

		v, err := build(next)
		if err != nil {
			return err
		}
		v.ReportID = reportID
		v.VersionNumber = next
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Report{}).
			Where("id = ?", reportID).
			Update("latest_published_version_id", v.ID).Error; err != nil {
			return err
		}
		saved = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}
