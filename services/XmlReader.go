package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	logger2 "github.com/Sterks/FReader/logger"
	"github.com/Sterks/Pp.Common/common"
	"github.com/Sterks/rXmlReader/config"
	"github.com/Sterks/rXmlReader/db"
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
}

//XMLReader ...
type XMLReader struct {
	config *config.Config
	logger *logger2.Logger
	db     *db.Database
}

//NewXMLReader ...
func NewXMLReader(config *config.Config) *XMLReader {
	return &XMLReader{
		config: config,
		db: &db.Database{},
	}
}

// OpenDatabase ...
func (x *XMLReader) Start(config *config.Config) *XMLReader {
	con := x.db.OpenDatabase(config)
	x.db.Db = con
	return x
}

func (x *XMLReader) UnzipFiles(msgs <-chan amqp.Delivery, forever chan bool) {
	go func() {
		for d := range msgs {
			// f := bytes.NewReader(d.Body)

			var inf InformationFile

			err := json.Unmarshal(d.Body, &inf)
			if err != nil {
				fmt.Println("Can't deserislize")
			}

			body := inf.FileZip

			zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
			if err != nil {
				log.Fatal("Не могу прочитать содержимое файла", err)
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
					fmt.Println(hash)
					zf.Close()


					zf2, err3 := zipFile.Open()

					id := x.db.LastID()

					os := zipFile.FileInfo()
					x.XMLSaver(id, zf2, os)
					x.db.CreateInfoFile(os, inf.Region, hash, inf.Fullpath, inf.FileID)

					//
					//unzippedFileBytes, err := readZipFile(zipFile)
					//if err != nil {
					//	log.Print("Перевод в байты", err)
					//	continue
					//}
					//_ = unzippedFileBytes // this is unzipped file bytes
				}
			}
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
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
func (x *XMLReader)  CreateFolder(config *config.Config, ident int) string {
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
	// if err := os.MkdirAll(lv4, 0755); err != nil {
	// 	log.Fatal(err)
	// }
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
