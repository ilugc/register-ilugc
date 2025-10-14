package register_ilugc

import (
	"context"
	"encoding/hex"
	"hash/fnv"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Db struct {
	Db *gorm.DB
}

type MParticipant struct {
	gorm.Model
	Chksum string
	Participant
}

func CreateDb() *Db {
	self := &Db{}
	return self
}

func (self *Db) Init() error {
	var err error
	self.Db, err = gorm.Open(sqlite.Open("participants.db"), &gorm.Config{})
	if err != nil {
		G.logger.Println(err)
		return err
	}
	if err = self.Db.AutoMigrate(&MParticipant{}); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Db) Chksum(participant *Participant) (string, error) {
	hash := fnv.New64()
	if _, err := hash.Write([]byte(participant.Name)); err != nil {
		G.logger.Println(err)
		return "", err
	}
	if _, err := hash.Write([]byte(participant.Email)); err != nil {
		G.logger.Println(err)
		return "", err
	}
	if _, err := hash.Write([]byte(participant.Mobile)); err != nil {
		G.logger.Println(err)
		return "", err
	}
	chksum := hex.EncodeToString(hash.Sum(nil))
	return chksum, nil
}

func (self *Db) Write(participant *Participant) error {
	chksum, err := self.Chksum(participant)
	if err != nil {
		G.logger.Println(err)
		return err
	}

	ctx := context.Background()
	count, err := gorm.G[MParticipant](self.Db).Where("chksum = ?", chksum).Count(ctx, "")
	if err != nil {
		G.logger.Println(err)
		return err
	}
	if count > 0 {
		return nil
	}

	if err := gorm.G[MParticipant](self.Db).Create(ctx, &MParticipant{Chksum: chksum, Participant: *participant}); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Db) Read(chksum string) (*Participant, error) {
	ctx := context.Background()
	mparticipant, err := gorm.G[MParticipant](self.Db).Where("chksum = ?", chksum).First(ctx)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}
	return &mparticipant.Participant, nil
}

func (self *Db) Delete(chksum string) error {
	ctx := context.Background()
	_, err := gorm.G[MParticipant](self.Db).Where("chksum = ?", chksum).Delete(ctx)
	if err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}
