package model

import (
	"errors"
	"github.com/zu1k/coolq-pushbot/core/config"
)

type Subscribe struct {
	ID                 uint `gorm:"primary_key;AUTO_INCREMENT"`
	UserID             int64
	SourceID           uint
	EnableNotification int
	EnableTelegraph    int
	EditTime
}

func RegistFeed(userID int64, sourceID uint) error {
	var subscribe Subscribe
	db := getConnect()
	defer db.Close()

	if err := db.Where("user_id=? and source_id=?", userID, sourceID).Find(&subscribe).Error; err != nil {
		if err.Error() == "record not found" {
			subscribe.UserID = userID
			subscribe.SourceID = sourceID
			subscribe.EnableNotification = 1
			subscribe.EnableTelegraph = 1
			err := db.Create(&subscribe).Error
			if err == nil {
				return nil
			}
		}
		return err
	}
	return nil
}

func GetSubscribeByUserIDAndSourceID(userID int64, sourceID uint) (*Subscribe, error) {
	db := getConnect()
	defer db.Close()
	var sub Subscribe
	db.Where("user_id=? and source_id=?", userID, sourceID).First(&sub)
	if sub.UserID != int64(userID) {
		return nil, errors.New("未订阅该RSS源")
	}
	return &sub, nil
}

func GetSubscribeByUserIDAndURL(userID int, url string) (*Subscribe, error) {
	db := getConnect()
	defer db.Close()
	var sub Subscribe
	source, err := GetSourceByUrl(url)
	if err != nil {
		return nil, err
	}
	db.Where("user_id=? and source_id=?", userID, source.ID).First(&sub)
	if sub.UserID != int64(userID) {
		return nil, errors.New("未订阅该RSS源")
	}
	return &sub, nil
}
func GetSubscriberBySource(s *Source) []Subscribe {
	if s == nil {
		return []Subscribe{}
	}

	db := getConnect()
	defer db.Close()
	var subs []Subscribe

	db.Where("source_id=?", s.ID).Find(&subs)
	return subs
}

func UnsubByUserIDAndSource(userID int64, source *Source) error {

	if source == nil {
		return errors.New("nil pointer")
	}

	db := getConnect()
	defer db.Close()
	var sub Subscribe
	db.Where("user_id=? and source_id=?", userID, source.ID).First(&sub)
	if sub.UserID != userID {
		return errors.New("未订阅该RSS源")
	}
	db.Delete(&sub)
	if source.GetSubscribeNum() < 1 {
		source.DeleteDueNoSubscriber()
	}
	return nil
}

func UnsubByUserIDAndSubID(userID int64, subID uint) error {
	db := getConnect()
	defer db.Close()

	var sub Subscribe
	db.Where("id=?", subID).First(&sub)

	if sub.UserID != userID {
		return errors.New("未找到该条订阅")
	}
	db.Delete(&sub)

	source, _ := GetSourceById(sub.SourceID)
	if source.GetSubscribeNum() < 1 {
		source.DeleteDueNoSubscriber()
	}
	return nil
}

func UnsubAllByUserID(userID int64) (success int, fail int, err error) {
	db := getConnect()
	defer db.Close()
	success = 0
	fail = 0
	var subs []Subscribe

	db.Where("user_id=?", userID).Find(&subs)

	for _, sub := range subs {
		err := sub.Unsub()
		if err != nil {
			fail += 1
		} else {
			success += 1
		}
	}
	err = nil

	return
}

func GetSubByUserIDAndURL(userID int64, url string) (*Subscribe, error) {
	db := getConnect()
	defer db.Close()
	var sub Subscribe
	source, err := GetSourceByUrl(url)
	if err != nil {
		return &sub, err
	}
	err = db.Where("user_id=? and source_id=?", userID, source.ID).First(&sub).Error
	return &sub, err
}

func GetSubsByUserID(userID int64) ([]Subscribe, error) {
	db := getConnect()
	defer db.Close()

	var subs []Subscribe

	db.Where("user_id=?", userID).Find(&subs)

	return subs, nil
}

func UnsubByUserIDAndSourceURL(userID int64, url string) error {
	source, err := GetSourceByUrl(url)
	if err != nil {
		return err
	}
	err = UnsubByUserIDAndSource(userID, source)
	return err
}

func GetSubscribeByID(id int) (*Subscribe, error) {
	db := getConnect()
	defer db.Close()
	var sub Subscribe
	db.Where("id=?  ", id).First(&sub)
	return &sub, nil
}

func (s *Subscribe) ToggleNotification() error {
	if s.EnableNotification != 1 {
		s.EnableNotification = 1
	} else {
		s.EnableNotification = 0
	}
	return nil
}

func (s *Subscribe) ToggleTelegraph() error {
	if s.EnableTelegraph != 1 {
		s.EnableTelegraph = 1
	} else {
		s.EnableTelegraph = 0
	}
	return nil
}

func (s *Source) ToggleEnabled() error {
	if s.ErrorCount >= config.ErrorThreshold {
		s.ErrorCount = 0
	} else {
		s.ErrorCount = config.ErrorThreshold
	}

	///TODO a hack for save source changes
	s.Save()

	return nil
}

func (s *Subscribe) Unsub() (err error) {
	if s.ID == 0 {
		return errors.New("can't delete 0 subscribe")
	}
	db := getConnect()
	defer db.Close()

	err = db.Delete(&s).Error
	return
}

func (s *Subscribe) Save() {
	db := getConnect()
	defer db.Close()
	db.Save(&s)
	return
}