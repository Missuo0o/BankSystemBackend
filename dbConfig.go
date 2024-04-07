package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Set struct {
	DBConf DBConf `yaml:"db"`
	// AppConf AppConf `yaml:"app"`
}

type DBConf struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func (conf *Set) getConf() *Set {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return conf
}

func connectDB() (*gorm.DB, error) {
	var set Set
	set.getConf()
	var dbConf = set.DBConf
	dsn := dbConf.User + ":" + dbConf.Password + "@tcp(" + dbConf.Host + ":" + dbConf.Port + ")/" + dbConf.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
