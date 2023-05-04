package main

import (
	"encoding/json"
	"fmt"
	appUCase "go-empty/pkg/app/app/usecase"
	platformRepo "go-empty/pkg/app/platform/repo/mysql"
	"go-empty/pkg/app/user/usecase"
	"go-empty/pkg/config"
	"go-empty/pkg/db"
	"go-empty/pkg/models"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gorm.io/gorm/logger"
)

const confPath = "./config.json"
const seedPath = "./seed.json"

func main() {

	cfg, err := config.New(confPath)
	if err != nil {
		log.Fatal(err)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	l := db.Logger{
		Logger: newLogger,
	}
	ddb, err := db.GormOpen(&cfg.DB, &l)
	if err != nil {
		log.Fatal(err)
	}
	db.MainDB = ddb

	fmt.Println("LoadSeed --> Start")

	err = saveSeed()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("LoadSeed --> Done")
}

type Seed struct {
	Users     []models.User
	Platforms []models.Platform
}

func saveSeed() error {

	fmt.Println("Seed: Create User --> Start")

	s, err := loadSeed(seedPath)
	if err != nil {
		return err
	}

	uIO := usecase.User{}
	for _, user := range s.Users {
		u, err := uIO.Register(&user) //char(32) will be error
		if err != nil {
			fmt.Printf("err =%s, for Register User: %+v", err.Error(), u)
		} else {
			fmt.Printf("Register User: %+v\n", u)
		}

	}
	fmt.Println("Seed: Create User  --> Done")
	fmt.Println("Seed: Create Platform --> Start")
	pIO := platformRepo.Platform{}
	for _, platform := range s.Platforms {
		u, err := pIO.CreatePlatform(&platform) //char(32) will be error
		if err != nil {
			fmt.Printf("err =%s, for Create Platform: %+v", err.Error(), u)
		} else {
			fmt.Printf("Create Platform: %+v\n", u)
		}

	}

	fmt.Println("Seed: Create Platform  --> Done")

	fmt.Println("Seed: Create App genre --> Start")

	appHandle := appUCase.AppHandle{}
	appGe := &models.AppGenre{
		Type:  "default",
		Brief: "default",
	}
	appHandle.RegGenre(appGe)

	fmt.Println("Seed: Create App genre  --> Done")

	return nil
}

func loadSeed(path string) (*Seed, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	type LoopUsers struct {
		models.User
		Count int `json:"count"`
	}
	type TmpSeed struct {
		Users     []models.User     `json:"Users"`
		LoopUsers []LoopUsers       `json:"LoopUsers"`
		Platforms []models.Platform `json:"Platforms"`
	}

	var tmpSeed TmpSeed

	err = json.Unmarshal(buf, &tmpSeed)
	if err != nil {
		return nil, err
	}

	seed := &Seed{}

	for _, v := range tmpSeed.LoopUsers {
		fmt.Println(v)
		for i := 0; i < v.Count; i++ {
			u := v.User
			no := fmt.Sprintf("%d", i+1)
			u.Name = u.Name + no
			u.Account = u.Account + no
			u.Password = u.Password + no
			tmpSeed.Users = append(tmpSeed.Users, u)
		}
	}

	seed.Users = tmpSeed.Users
	seed.Platforms = tmpSeed.Platforms

	return seed, nil
}
