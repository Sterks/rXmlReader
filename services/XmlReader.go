package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sterks/Pp.Common.Db/db"
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
	x.db.OpenDatabase()
	return x
}

func (x *XMLReader) UnzipFiles(msgs <-chan amqp.Delivery, forever chan bool, config *config.Config) {
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
				f, err2 := os.Create("logFile")
				g := fmt.Sprintf("Не могу прочитать файл - %v - %v", err, inf.NameFile)
				l := strings.NewReader(g)
				_, _ = io.Copy(f, l)
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
					defer zf.Close()
					hash := common.GetHash(zf)
					fmt.Println(hash)
					_ = zf.Close()

					zf2, err3 := zipFile.Open()
					defer zf2.Close()
					zf3, err4 := zipFile.Open()
					if err4 != nil {
						log.Println("Не могу прочитать - %v", err4)
					}
					defer zf2.Close()
					id := x.db.LastID()
					ost := zipFile.FileInfo()
					x.XMLSaver(id, zf2, ost)

					zzz, err5 := ioutil.ReadAll(zf3)
					if err5 != nil {
						log.Println(err5)
					}
					if inf.TypeFile == "notifications44" {
						x.amq.PublishSend(config, ost, "Notifications44OpenFile", zzz, id, ost.Name(), inf.Fullpath)
					} else if inf.TypeFile == "protocols44" {
						x.amq.PublishSend(config, ost, "Protocols44OpenFile", zipFile.Extra, id, ost.Name(), inf.Fullpath)
					} else {
					pattern := ".+Notice"
					reg := value.Name()
					matched, err := regexp.MatchString(pattern, reg)
					if err != nil {
						log.Printf("Не могу распознать слова - %v", err)
					}
					if matched {
						x.amq.PublishSend(config, ost, "Notifications223OpenFile", zipFile.Extra, id, ost.Name(), inf.Fullpath)
					} else {
					pattern := ".+Protocol"
					reg := value.Name()
					matched, err := regexp.MatchString(pattern, reg)
					if err != nil {
						log.Printf("Не могу распознать слова - %v", err)
					}
					if matched {
							x.amq.PublishSend(config, ost, "Protocols223OpenFile", zipFile.Extra, id, ost.Name(), inf.Fullpath)
					} else {
						x.amq.PublishSend(config, ost, "Not choose", zipFile.Extra, id, ost.Name(), inf.Fullpath)
					}
					x.db.CreateInfoFile(ost, inf.Region, hash, inf.Fullpath, inf.TypeFile, inf.TypeFile)
				}
				}
				}
			}
		}
	}()

	log.Printf("[*] Ждем сообщений. Для вызода нажмите CTRL+C")
	<-forever
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
