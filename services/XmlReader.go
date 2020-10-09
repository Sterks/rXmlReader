package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Sterks/Pp.Common.Db/db"
	"github.com/Sterks/Pp.Common.Db/models"
	"github.com/Sterks/rXmlReader/rabbit"
	"github.com/streadway/amqp"

	"github.com/Sterks/Pp.Common/common"
	logger2 "github.com/Sterks/fReader/logger"
	"github.com/Sterks/rXmlReader/config"
	log "github.com/sirupsen/logrus"
)

// InformationFile Структура для Rabbit
type InformationFile struct {
	FileID   int
	NameFile string
	SizeFile int64
	DateMode time.Time
	Fullpath string
	Region   string
	FileZip  []byte
	TypeFile string
}

//XMLReader ...
type XMLReader struct {
	config *config.Config
	logger *logger2.Logger
	db     *db.Database
	amq    *rabbit.ProducerMQ
}

//NewXMLReader ...
func NewXMLReader(config *config.Config) *XMLReader {
	return &XMLReader{
		config: config,
		db:     &db.Database{},
		amq:    &rabbit.ProducerMQ{},
	}
}

// OpenDatabase ...
func (x *XMLReader) Start(config *config.Config) *XMLReader {
	x.db.OpenDatabase(x.config.Postgres.Host, x.config.Postgres.Port, x.config.Postgres.User, x.config.Postgres.Password, x.config.Postgres.DBName)
	return x
}

// UnzipFiles ...
func (x *XMLReader) UnzipFiles(msgs <-chan amqp.Delivery, forever chan bool, config *config.Config, typeFile string) {
	go func() {
		for d := range msgs {
			var inf InformationFile
			err := json.Unmarshal(d.Body, &inf)
			if err != nil {
				fmt.Println("Can't deserislize")
			}

			body := inf.FileZip

			zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
			if err != nil {
				// f, err2 := os.Create("logFile")
				f, err2 := os.OpenFile("logFile", os.O_APPEND|os.O_WRONLY, 0644)
				defer f.Close()
				g := fmt.Sprintf("Не могу прочитать файл - %v - %v\n", err, inf.NameFile)
				f.WriteString(g)
				log.Printf("Не могу прочитать содержимое файла - %v", err)
				if err2 != nil {
					log.Printf("Не могу записать в лог - %v", err)
				}
				continue
			}

			// Read all the files from zip archive
			for _, zipFile := range zipReader.File {
				res := strings.Contains(zipFile.Name, ".xml")
				if res == true {
					fmt.Println("Reading file:", zipFile.Name+" - "+inf.NameFile, inf.SizeFile)
					zf, err3 := zipFile.Open()
					if err3 != nil {
						log.Printf("Не могу прочитать файл: %v", err3)
						continue
					}

					hash := common.GetHash(zf)
					zf2, err3 := zipFile.Open()

					id := x.db.LastID()
					ost := zipFile.FileInfo()
					x.XMLSaver(id, zf2, ost)

					zf3, err4 := zipFile.Open()
					if err4 != nil {
						log.Println("Не могу прочитать - %v", err4)
					}

					zzz, err5 := ioutil.ReadAll(zf3)
					if err5 != nil {
						log.Println(err5)
					}

					if inf.TypeFile == "notifications44" {
						x.amq.PublishSend(config, ost, "Notifications44OpenFile", zzz, id, ost.Name(), inf.Fullpath)
						typeFile = x.AddToDatabase(ost, typeFile, inf, hash)
					} else if inf.TypeFile == "protocols44" {
						x.amq.PublishSend(config, ost, "Protocols44OpenFile", zzz, id, ost.Name(), inf.Fullpath)
						typeFile = x.AddToDatabase(ost, typeFile, inf, hash)
					} else if inf.TypeFile == "notifications223" {
						typeFile = x.AddToDatabase(ost, typeFile, inf, hash)
						x.amq.PublishSend(config, ost, "Notifications223OpenFile", zipFile.Extra, id, ost.Name(), inf.Fullpath)
					} else if inf.TypeFile == "protocols223" {
						x.amq.PublishSend(config, ost, "Protocols223OpenFile", zipFile.Extra, id, ost.Name(), inf.Fullpath)
						typeFile = x.AddToDatabase(ost, typeFile, inf, hash)
					} else {
						x.amq.PublishSend(config, ost, "Not choose", zipFile.Extra, id, ost.Name(), inf.Fullpath)
						typeFile = x.AddToDatabase(ost, typeFile, inf, hash)
					}
				}
			}
		}
	}()

	log.Printf("[*] Ждем сообщений в %s очередь. Для вызода нажмите CTRL+C", typeFile)
	<-forever
}

// AddToDatabase Добавляет запись файла
func (x *XMLReader) AddToDatabase(ost os.FileInfo, typeFile string, inf InformationFile, hash string) string {
	ext := filepath.Ext(ost.Name())

	var fileType models.FileType
	x.db.Database.Table("FilesTypes").Where("ft_ext = ?", ext).Find(&fileType)

	var sr models.SourceResources
	typeFile = strings.ToLower(typeFile)
	if err := x.db.Database.Table("SourceResources").Where("sr_name = ?", typeFile).Find(&sr).Error; err != nil {
		log.Fatalf("Не могу определить Resource - %v", err)
	}

	var gf models.SourceRegions
	x.db.Database.Table("SourceRegions").Where("r_name = ?", inf.Region).Find(&gf)

	var f models.File
	f.TName = ost.Name()
	f.TArea = gf.RID
	f.TType = fileType.FTID
	f.THash = hash
	f.TSize = ost.Size()
	f.CreatedAt = time.Now()
	f.TDateCreateFromSource = ost.ModTime()
	f.TDateLastCheck = time.Now()
	f.TFullpath = inf.Fullpath
	f.TSourceResources = sr.SRID
	x.db.Database.Table("Files").Create(&f)
	return typeFile
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func (x *XMLReader) XMLSaver(id int, zf2 io.ReadCloser, info os.FileInfo) {
	if err := os.MkdirAll("./Files", 0755); err != nil {
		log.Errorf("Не могу создать директорию - %v\n", err)
	}
	pathLocal := x.CreateFolder(x.config, id)
	nameFile := GenerateID(id)
	file, _ := os.Create(x.config.Directory.MainFolder + "/" + pathLocal + nameFile)
	defer file.Close()
	io.Copy(file, zf2)
}

// CreateFolder ...
func (x *XMLReader) CreateFolder(config *config.Config, ident int) string {
	saveDir := config.Directory.MainFolder
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		log.Errorf("Не могу создать директорию - %v\n", err)
	}
	stringID := GenerateID(ident)
	lv1 := fmt.Sprint(stringID[0:3])
	lv2 := fmt.Sprint(stringID[3:6])
	lv3 := fmt.Sprint(stringID[6:9])
	// lv4 := fmt.Sprintln(stringID[9:12])
	if err := os.MkdirAll(saveDir+"/"+lv1, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(saveDir+"/"+lv1+"/"+lv2, 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(saveDir+"/"+lv1+"/"+lv2+"/"+lv3, 0755); err != nil {
		log.Fatal(err)
	}
	path := fmt.Sprintf("%s/%s/%s/", lv1, lv2, lv3)
	return path
}

// GenerateID - Герерация строки длинной 12 символов
func GenerateID(ident int) string {
	ident = ident + 1
	word := strconv.Itoa(ident)
	ch := len(word)
	nool := 12 - ch
	var ap string
	ap = word
	for i := 0; i < nool; i++ {
		ap = fmt.Sprintf("0%s", ap)
	}
	return ap
}
