// Definition of the structures and SQL interaction functions
package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	usermodels "github.com/quorumsco/oauth2/models"
)

// TeamSQL contains a Gorm client and the team and gorm related methods
type TeamSQL struct {
	DB *gorm.DB
}

// Save inserts a new team into the database
func (s *TeamSQL) Save(t *Team, args TeamArgs) error {
	if t == nil {
		return errors.New("save: team is nil")
	}

	var c = &usermodels.User{ID: args.UserID}

	if t.ID == 0 {
		err := s.DB.Debug().Model(c).Association("Teams").Append(t).Error
		s.DB.Last(t)
		return err
	}

	return s.DB.Model(c).Association("Teams").Replace(t).Error
}

// Delete removes a team from the database
func (s *TeamSQL) Delete(t *Team, args TeamArgs) error {
	return s.DB.Model(&usermodels.User{ID: args.UserID}).Association("Teams").Delete(t).Error
}

// Find return all the teams containing a given groupID from the database
func (s *TeamSQL) Find(args TeamArgs) ([]Team, error) {
	var (
		teams []Team
		c     = &usermodels.User{ID: args.UserID}
	)

	err := s.DB.Model(c).Association("Teams").Find(&teams).Error
	if err != nil {
		return nil, err
	}

	return teams, nil
}
