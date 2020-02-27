package db

import (
	"fmt"
	model "github.com/Sterks/fReader/models"
	"os"
	"path/filepath"
	"time"

	"github.com/Sterks/fReader/logger"
	"github.com/Sterks/rXmlReader/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" //....
	log "github.com/sirupsen/logrus"
)

//Database ...
type Database struct {
	Config *config.Config
	Db     *gorm.DB
	Logger *logger.Logger
}

const (
	host     = "localhost"
	port     = 5432
	user     = "user_ro"
	password = "4r2w3e1q"
	dbname   = "freader"
)

// NewDatabase ...
func NewDataBase(config *config.Config) *Database {
	return &Database{
		Config: nil,
		Db:     nil,
		Logger: nil,
	}
}

func (d *Database) OpenDatabase(config *config.Config) *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	dbd, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		log.Errorf("Соединиться не удалось - %s", err)
	}
	if err2 := dbd.DB().Ping(); err2 != nil {
		log.Errorf("База не отвечает", err2)
	}
	return dbd
}

func (d *Database) LastID() int {
	var ff model.File
	d.Db.Table("Files").Last(&ff)
	return ff.TID
}

func (d *Database) CreateInfoFile(info os.FileInfo, region string, hash string, fullpath string, id int) int {
	d.Db.LogMode(true)

	var gf model.SourceRegions
	d.Db.Table("SourceRegions").Where("r_name = ?", region).Find(&gf)

	checker := d.CheckExistFileDb(info, hash)
	if checker != 0 {
		var lf model.File
		d.Db.Table("Files").Where("f_id = ?", checker).Find(&lf)
		lf.TDateLastCheck = time.Now()
		d.Db.Save(&lf)
		log.Printf("Дата успешно обновлена", lf.TDateLastCheck.String())
	}
	if checker == 0 {

		var fileType model.FileType
		var lastID model.File
		ext := filepath.Ext(info.Name())
		d.Db.Table("FilesTypes").Where("ft_ext = ?", ext).Find(&fileType)

		d.Db.Table("Files")
		d.Db.Create(&model.File{
			TName:                 info.Name(),
			TParent:               id,
			TArea:                 gf.RID,
			FileType:              fileType,
			TType:                 fileType.FTID,
			THash:                 hash,
			TSize:                 info.Size(),
			CreatedAt:             time.Now(),
			TDateCreateFromSource: info.ModTime(),
			TDateLastCheck:        time.Now(),
			TFullpath:             fullpath,
		}).Scan(&lastID)
		log.Printf("Файл успешно добавлен - %v", lastID.TName)
		return lastID.TID
	} else {
		fmt.Printf("Файл существует - %v\n", info.Name())
		return 0
	}
}

// CheckExistFileDb ...
func (d *Database) CheckExistFileDb(file os.FileInfo, hash string) int {
	var ff model.File
	d.Db.Table("Files").Where("f_hash = ? and f_size = ? and f_name = ?", hash, file.Size(), file.Name()).Find(&ff)
	return ff.TID
}

//
//// CreateInfoFile ...
//func (d *Database) CreateInfoFile(info os.FileInfo, region string, hash string, fullpath string) int {
//	// d.DatabaseGorm.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false).Create(&files)
//	// filesTypes := d.DatabaseGorm.Table("FileType")
//	d.DatabaseGorm.LogMode(true)
//
//	var gf model.SourceRegions
//	d.DatabaseGorm.Table("SourceRegions").Where("r_name = ?", region).Find(&gf)
//
//	checker := d.CheckExistFileDb(info, hash)
//	if checker != 0 {
//		var lf model.File
//		d.DatabaseGorm.Table("Files").Where("f_id = ?", checker).Find(&lf)
//		lf.TDateLastCheck = time.Now()
//		d.DatabaseGorm.Save(&lf)
//		log.Printf("Дата успешно обновлена", lf.TDateLastCheck.String())
//	}
//	if checker == 0 {
//
//		var fileType model.FileType
//		var lastID model.File
//		d.DatabaseGorm.Table("FilesTypes").Where("ft_name = ?", "ZIP архив").Find(&fileType)
//
//		d.DatabaseGorm.Table("Files")
//		d.DatabaseGorm.Create(&model.File{
//			TName:                 info.Name(),
//			TArea:                 gf.RID,
//			FileType:              fileType,
//			TType:                 fileType.FTID,
//			THash:                 hash,
//			TSize:                 info.Size(),
//			CreatedAt:             time.Now(),
//			TDateCreateFromSource: info.ModTime(),
//			TDateLastCheck:        time.Now(),
//			TFullpath:             fullpath,
//		}).Scan(&lastID)
//		log.Printf("Файл успешно добавлен - ", lastID.TName)
//		return lastID.TID
//	} else {
//		fmt.Printf("Файл существует - %v\n", info.Name())
//		return 0
//	}
//}
//
////LastID ...
//func (d *Database) LastID() int {
//	var ff model.File
//	d.DatabaseGorm.Table("Files").Last(&ff)
//	return ff.TID
//}
//
//// CheckerExistFileDBNotHash ...
//func (d *Database) CheckerExistFileDBNotHash(file os.FileInfo) (int, string) {
//	var ff model.File
//	fmt.Printf("%v - %v - %v", file.Size(), file.Name(), file.ModTime())
//	d.DatabaseGorm.Table("Files").Where("f_size = ? and f_name = ? and f_date_create_from_source = ?", file.Size(), file.Name(), file.ModTime()).Find(&ff)
//	return ff.TID, ff.THash
//}
//
//// CheckExistFileDb ...
//func (d *Database) CheckExistFileDb(file os.FileInfo, hash string) int {
//	var ff model.File
//	d.DatabaseGorm.Table("Files").Where("f_hash = ? and f_size = ? and f_name = ?", hash, file.Size(), file.Name()).Find(&ff)
//	return ff.TID
//}
//
////CheckRegionsDb Проверка существует ли регион в базе данных
//func (d *Database) CheckRegionsDb(region string) int {
//	var reg model.SourceRegions
//	d.DatabaseGorm.Table("SourceRegions").Where("r_name = ?", region).First(&reg)
//	return reg.RID
//}
//
////ReaderRegionsDb Все регионы из базы
//func (d *Database) ReaderRegionsDb() []model.SourceRegions {
//	var regions []model.SourceRegions
//	d.DatabaseGorm.Table("SourceRegions").Find(&regions)
//	return regions
//}
//
////AddRegionsDb ...
//func (d *Database) AddRegionsDb(region string) {
//	var reg model.SourceRegions
//	reg.RName = region
//	reg.RDateCreate = time.Now()
//	reg.RDateUpdate = time.Now()
//	d.DatabaseGorm.Table("SourceRegions").Create(&reg)
//}
//
//// FirstOrCreate Создать или получить
//func (d *Database) FirstOrCreate(region string) model.SourceRegions {
//	var reg model.SourceRegions
//	reg.RName = region
//	reg.RDateCreate = time.Now()
//	reg.RDateUpdate = time.Now()
//	d.DatabaseGorm.Table("SourceRegions").Where("r_name = ?", region).FirstOrCreate(&reg)
//	return reg
//}
