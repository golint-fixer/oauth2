// Definition of the structures and SQL interaction functions
package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/quorumsco/logs"
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

	//var c = usermodels.Team{ID: args.UserID}

	if t.ID == 0 {
		//err := s.DB.Debug().Model(c).Association("Teams").Append(t).Error
		err := s.DB.Debug().Model(t).Association("Users").Append(t).Error
		s.DB.Last(t)
		return err
	}

	return s.DB.Model(t).Association("Users").Replace(t).Error
}

// Delete removes a team from the database
func (s *TeamSQL) Delete(t *Team, args TeamArgs) error {
	return s.DB.Model(&usermodels.User{ID: args.UserID}).Association("Teams").Delete(t).Error
}

// Find return all the teams containing a given groupID from the database
func (s *TeamSQL) Find(args TeamArgs) ([]Team, error) {

	var (
		teams []Team
		team  = Team{ID: args.TeamID}
		users []usermodels.User
		//user = usermodels.User{ID: 1}
		//c     = Team{GroupID: args.GroupID}
		//c = &usermodels.User{ID: 3}
	)
	//logs.Debug(c)
	logs.Debug(teams)
	err := s.DB.Debug().Model(&team).Related(&users, "Users").Error
	err = s.DB.Debug().Preload("Users").Find(&teams).Error
	//// SELECT * FROM "languages" INNER JOIN "user_languages" ON "user_languages"."language_id" = "languages"."id" WHERE "user_languages"."user_id" = 111
	//err = s.DB.Debug().Model(&team).Find(&teams).Error
	//err := s.DB.Debug().Model(&c).Association("Users").Find(&teams).Error
	//err := s.DB.Debug().Model(&c).Association("Users").Find(&teams).Error
	if err != nil {
		return nil, err
	}

	return teams, nil
}

/*
type User struct {
	ID             int64      `json:"id";gorm:"primary_key"`
	Mail           *string    `sql:"not null;unique" json:"mail"`
	Password       *string    `sql:"not null" json:"password"`
	Firstname      *string    `sql:"not null" json:"firstname"`
	Surname        *string    `sql:"not null" json:"surname"`
	Role           *string    `json:"role"`
	Cause          *string    `sql:"not null" json:"cause"`
	GroupID        uint       `json:"group_id"`
	OldgroupID     uint       `json:"oldgroup_id"`
	Validationcode *string    `json:"validationcode"`
	Phone          *string    `json:"phone"`
	Address        *string    `json:"address"`
	Created        *time.Time `json:"created,omitempty"`
}
*/

// func (s *TagSQL) Find(args TagArgs) ([]Tag, error) {
// 	var (
// 		tags []Tag
// 		c    = &Contact{ID: args.ContactID}
// 	)

// 	err := s.DB.Model(c).Association("Tags").Find(&tags).Error
// 	if err != nil {
// 		return nil, err
// 	}

// 	return tags, nil
// }
