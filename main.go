package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var (
	videoFile     []string // 视频文件
	fileNum       int      // 视频文件数量
	totalDuration uint32   // 总时长
)

func main() {
	fmt.Println("请输入视频文件夹地址:")
	var foldpath string
	reader := bufio.NewReader(os.Stdin)

	res, _, err := reader.ReadLine()

	if err != nil {
		panic(err)
	}

	// foldpath = "D:\\1_liuxiaobo\\download_baidu\\03-阶段三 C++核心编程和桌面应用开发"
	foldpath = string(res)
	fmt.Println("目录:", foldpath)

	// foldpath改成双斜杠
	foldpath = strings.Replace(foldpath, "\\", "\\\\", -1)

	// 判断文件夹是否存在
	if _, err := os.Stat(foldpath); os.IsNotExist(err) {
		fmt.Println("文件夹不存在")
		return
	}

	fmt.Println("正在查找视频文件...")
	videoList := getVideoFile(foldpath)
	fmt.Println("视频文件数量:", fileNum)

	// 已处理视频个数
	var videoOkNum int
	for _, v := range videoList {
		duration := GetMP4Duration(v)
		if err != nil {
			panic(err)
		}

		videoOkNum++

		// datetime转换为秒
		// fmt.Println("视频时长:", duration)
		if duration == 0 {
			fmt.Println("视频时长为0, 跳过")
			continue
		}
		totalDuration += duration

		fmt.Printf("视频总数: %d, 已处理: %d \n", len(videoList), videoOkNum)
		// fmt.Printf("当前时长为1: %s ------ \n", IntToTime(int(totalDuration)))
		// fmt.Printf("当前时长为2: %d ------ \n", totalDuration)
	}
	fmt.Println("总时长为:", IntToTime(int(totalDuration)))
}

// 递归查询文件夹下的视频文件
func getVideoFile(foldpath string) []string {
	// 获取文件夹
	files, err := ioutil.ReadDir(foldpath)
	if err != nil {
		panic(err)
	}
	// 遍历文件夹
	for _, file := range files {
		// 判断是否是文件夹, 直到找到视频文件
		if file.IsDir() {
			// 获取文件夹下的文件
			getVideoFile(foldpath + "\\" + file.Name())
		} else {
			// 声明视频格式的数组
			videoFormat := []string{".mp4", ".avi", ".rmvb", ".mkv", ".flv", ".wmv", ".mov", ".mpg", ".mpeg", ".3gp", ".dat", ".ts", ".rm", ".asf", ".ram", ".vob", ".m4v", ".f4v", ".f4p", ".f4a", ".f4b"}
			// 获取文件后缀
			fileSuffix := strings.ToLower(file.Name()[strings.LastIndex(file.Name(), "."):])
			// 判断文件后缀是否是视频格式
			for _, format := range videoFormat {
				if fileSuffix == format {
					videoFile = append(videoFile, foldpath+"\\"+file.Name())
					fileNum++
					// fmt.Println("视频文件:", foldpath+"\\"+file.Name())
					// fmt.Println("已找到", fileNum, "个视频文件")
				}
			}
		}
	}
	return videoFile
}

// BoxHeader 信息头
type BoxHeader struct {
	Size       uint32
	FourccType [4]byte
	Size64     uint64
}

//filePath 视频地址
func GetMP4Duration(filePath string) uint32 {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	var (
		info      = make([]byte, 0x10)
		boxHeader BoxHeader
		offset    int64 = 0
	)
	// 获取结构偏移
	for {
		_, err = file.ReadAt(info, offset)
		if err != nil {
			return 0
		}
		boxHeader = getHeaderBoxInfo(info)
		fourccType := getFourccType(boxHeader)
		if fourccType == "moov" {
			break
		}
		// 有一部分mp4 mdat尺寸过大需要特殊处理
		if fourccType == "mdat" {
			if boxHeader.Size == 1 {
				offset += int64(boxHeader.Size64)
				continue
			}
		}
		offset += int64(boxHeader.Size)
	}
	// 获取move结构开头一部分
	moveStartBytes := make([]byte, 0x100)
	_, err = file.ReadAt(moveStartBytes, offset)
	if err != nil {
		return 0
	}
	// 定义timeScale与Duration偏移
	timeScaleOffset := 0x1C
	durationOffset := 0x20
	timeScale := binary.BigEndian.Uint32(moveStartBytes[timeScaleOffset : timeScaleOffset+4])
	Duration := binary.BigEndian.Uint32(moveStartBytes[durationOffset : durationOffset+4])
	return Duration / timeScale
}

// getHeaderBoxInfo 获取头信息
func getHeaderBoxInfo(data []byte) (boxHeader BoxHeader) {
	buf := bytes.NewBuffer(data)
	_ = binary.Read(buf, binary.BigEndian, &boxHeader)
	return
}

// getFourccType 获取信息头类型
func getFourccType(boxHeader BoxHeader) (fourccType string) {
	fourccType = string(boxHeader.FourccType[:])
	return
}

// int转时分秒
func IntToTime(second int) string {
	hour := second / 3600
	minute := second % 3600 / 60
	second = second % 60
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

// 代码参考： https://blog.csdn.net/weixin_42141510/article/details/121513683
