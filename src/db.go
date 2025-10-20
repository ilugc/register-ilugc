package register

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"hash/fnv"
	"sort"

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

type MConfigDetails struct {
	gorm.Model
	ConfigDetails
}

func CreateDb() *Db {
	self := &Db{}
	return self
}

func (self *Db) Init() error {
	var err error
	self.Db, err = gorm.Open(sqlite.Open("register.db"), &gorm.Config{})
	if err != nil {
		G.logger.Println(err)
		return err
	}
	if err = self.Db.AutoMigrate(&MParticipant{}); err != nil {
		G.logger.Println(err)
		return err
	}
	if err = self.Db.AutoMigrate(&MConfigDetails{}); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Db) ParticipantChksum(participant *Participant) (string, error) {
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

func (self *Db) ParticipantWrite(participant *Participant) error {
	chksum, err := self.ParticipantChksum(participant)
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

func (self *Db) ParticipantRead(chksum string) (*Participant, error) {
	if len(chksum) <= 0 {
		err := errors.New("Empty chksum")
		G.logger.Println(err)
		return nil, err
	}

	ctx := context.Background()
	mparticipant, err := gorm.G[MParticipant](self.Db).Where("chksum = ?", chksum).First(ctx)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}
	return &mparticipant.Participant, nil
}

func (self *Db) ParticipantDelete(chksum string) error {
	if len(chksum) <= 0 {
		err := errors.New("Empty chksum")
		G.logger.Println(err)
		return err
	}

	ctx := context.Background()
	_, err := gorm.G[MParticipant](self.Db).Where("chksum = ?", chksum).Delete(ctx)
	if err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Db) ParticipantCount() (int64, error) {
	ctx := context.Background()
	count, err := gorm.G[MParticipant](self.Db).Count(ctx, "")
	if err != nil {
		G.logger.Println(err)
		return 0, err
	}
	return count, nil
}

func (self *Db) ParticipantCsv() ([]byte, error) {
	ctx := context.Background()
	mparticipants, err := gorm.G[MParticipant](self.Db).Find(ctx)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}

	csvbuffer := bytes.NewBuffer(nil)
	csvwriter := csv.NewWriter(csvbuffer)
	headers := []string{}
	for _, mparticipant := range mparticipants {
		participant := &mparticipant.Participant
		participantmap := StructToMap(participant)
		if len(headers) <= 0 {
			for header, _ := range participantmap {
				headers = append(headers, header)
			}
			sort.Strings(headers)
			csvwriter.Write(headers)
		}
		values := []string{}
		for _, header := range headers {
			values = append(values, participantmap[header])
		}
		csvwriter.Write(values)
		csvwriter.Flush()
	}
	return csvbuffer.Bytes(), nil
}

func (self *Db) ConfigDetailsWrite(configdetails *ConfigDetails) error {
	ctx := context.Background()
	_, err := gorm.G[MConfigDetails](self.Db).Where("1 = 1").Delete(ctx)
	if err != nil {
		G.logger.Println(err)
		return err
	}
	if err := gorm.G[MConfigDetails](self.Db).Create(ctx, &MConfigDetails{ConfigDetails: *configdetails}); err != nil {
		G.logger.Println(err)
		return err
	}
	return nil
}

func (self *Db) ConfigDetailsRead() (*ConfigDetails, error) {
	ctx := context.Background()
	mconfigdetails, err := gorm.G[MConfigDetails](self.Db).First(ctx)
	if err != nil {
		G.logger.Println(err)
		return nil, err
	}
	return &mconfigdetails.ConfigDetails, nil
}
