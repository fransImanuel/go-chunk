package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func main() {

	// // fileToBeChunked := "./somebigfile"
	fileToBeChunked := "./18sec_sample_vids.mp4"
	// fileToBeChunked := "./1hr_sample_vids.mp4"
	// fileToBeChunked := "./2hr_sample_vids.mp4"
	// fileToBeChunked := "c:\\Users\\user\\Download\\1hr_sample_vids.mp4"

	file, err := os.Open(fileToBeChunked)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	fileInfo, _ := file.Stat()

	var fileSize int64 = fileInfo.Size()

	// const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement
	const fileChunk = 5 * (1 << 20) // 10 MB, change this to your requirement

	// calculate total number of parts the file will be chunked into

	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	var removechunkname []string
	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		fmt.Printf("partSize: %v\n", partSize)
		// fmt.Printf("partSize: %v, partBuffer: %v.\n", partSize, partBuffer)

		file.Read(partBuffer)

		// write to disk
		fileName := "chunk_" + "400009" + strconv.FormatUint(i+1, 10)
		removechunkname = append(removechunkname, fileName)
		_, err := os.Create(fileName)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// write/save buffer to disk
		ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)

		fmt.Println("Split to : ", fileName)

		SendChunk(fileName, i+1, totalPartsNum)
	}

	ClearChunk(removechunkname)

	// // SendChunk()

	// GenerateMD5CheckSum("asdf12345")
}

func SendChunk(fileName string, i, totalIndex uint64) {
	// var data models.Purchasing1688Data
	var data interface{}

	fileDir, _ := os.Getwd()
	// fileName := "1hr_sample_vids.mp4"
	filePath := path.Join(fileDir, fileName)

	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// fmt.Println("filepath.Base(file.Name()): ", filepath.Base(file.Name()))
	part, err := writer.CreateFormFile("data", filepath.Base(file.Name()))
	if err != nil {
		panic(err)
	}
	io.Copy(part, file)
	writer.Close()

	// url := "http://192.168.15.21:8787/upload/refund-videos"
	url := "http://localhost:8787/upload/refund-videos"
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json")
	// req.Header.Set("Content-Type", "application/json")

	fmt.Println("---")
	fmt.Println(fileName)
	fmt.Println(i, totalIndex)
	fmt.Println("---")
	req.Header.Add("Content-Type", writer.FormDataContentType())
	// req.Header.Set("Content-Type", "application/octet-stream")
	// reqBody := `{
	// 	"id_so" : "testidso_18sec_vids",
	// 	"num_of_vid" : ` + "1" + `,
	// 	"curr_index" : ` + fmt.Sprintf("%v", i) + `,
	// 	"total_chunk" : ` + fmt.Sprintf("%v", totalIndex) + `,
	// 	"md5" : "hashrandom",
	// 	"type" : "mp4",
	// }`

	// body := bytes.NewBuffer([]byte(reqBody))

	req.Header.Set("curr_index", fmt.Sprintf("%v", i))
	req.Header.Set("total_chunk", fmt.Sprintf("%v", totalIndex))
	req.Header.Set("id_so", "400009")
	req.Header.Set("type", "mp4")
	req.Header.Set("vid_no", "1")
	req.Header.Set("md5", "FDdPlF3PV4l51M+GXRo2kg==")
	// req.Header.Set("md5", "16d4c0c84d0c26bb877446d59fa056b3") // 1 jam

	fmt.Println(req.Header.Clone())

	// client := &http.Client{}
	// res, err := client.Do(req)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(res.Status)

	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(" error ", url)
	}

	defer resp2.Body.Close()
	bodyres, _ := ioutil.ReadAll(resp2.Body)
	json.Unmarshal([]byte(bodyres), &data)
	// fmt.Printf(resp2)
	// fmt.Printf("%v", []byte(bodyres))
	fmt.Println(string(bodyres))
	fmt.Println(url)
	// return data
}

func ClearChunk(f []string) {
	//delete chunk
	for _, v := range f {
		if err := os.Remove(v); err != nil {
			fmt.Println("os.remove:", err)
		}
	}

}

func GenerateMD5CheckSum(id_so string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	wd += "/refund-videos/" + id_so + ".mp4"
	// wd += "\\test.txt"
	fmt.Println(wd)
	file, err := os.Open(wd)

	if err != nil {
		return "", err
	}

	defer file.Close()

	hash := md5.New()
	// hash.Write([]byte(id_so))
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	// fmt.Printf("%s MD5 checksum is %x \n", file.Name(), hash.Sum(nil))
	// data := []byte(id_so)
	bodyHash := hash.Sum(nil)
	bodyHashBase64 := base64.StdEncoding.EncodeToString(bodyHash)

	fmt.Println(bodyHashBase64)

	return bodyHashBase64, nil
}
