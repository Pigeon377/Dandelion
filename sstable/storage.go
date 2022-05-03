package sstable

import (
	"Dandelion/skiplist"
	"Dandelion/util"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	sep                    = util.Sep
	defaultBufferSize      = 2 * 4096
	closingBound           = defaultBufferSize * 9 / 10
	dataFilePrefix         = "dandelion_db_storage_data_"
	indexFilePrefix        = "dandelion_db_storage_index_"
	storageDBFileDirectory = "data"

	level0PathPrefix = "level0" + string(os.PathSeparator)
	level1PathPrefix = "level1" + string(os.PathSeparator)
	level2PathPrefix = "level2" + string(os.PathSeparator)
	level3PathPrefix = "level3" + string(os.PathSeparator)
	level4PathPrefix = "level4" + string(os.PathSeparator)
	level5PathPrefix = "level5" + string(os.PathSeparator)

	storageFilePathPrefix = storageDBFileDirectory + string(os.PathSeparator)
	level0FilePathPrefix  = storageFilePathPrefix + level0PathPrefix
	level1FilePathPrefix  = storageFilePathPrefix + level1PathPrefix
	level2FilePathPrefix  = storageFilePathPrefix + level2PathPrefix
	level3FilePathPrefix  = storageFilePathPrefix + level3PathPrefix
	level4FilePathPrefix  = storageFilePathPrefix + level4PathPrefix
	level5FilePathPrefix  = storageFilePathPrefix + level5PathPrefix
	level1MaxSize         = 1024 * 1024 * 8
	//every indexRangeSize element generator a index
	indexRangeSize = 32
)

var (
	currentProjectPath, _ = os.Getwd()
	currentStoragePath    = currentProjectPath + string(os.PathSeparator) + storageDBFileDirectory + string(os.PathSeparator)

	levelPrefixArray = [6]string{
		level0PathPrefix,
		level1PathPrefix,
		level2PathPrefix,
		level3PathPrefix,
		level4PathPrefix,
		level5PathPrefix,
	}
)

func writeLevel0DBToFile(suffix string, kv []*util.KV) error {
	return writeDBToFile(suffix, kv, 0)
}

func writeDBToFile(suffix string, kv []*util.KV, level int) error {

	file, err := os.OpenFile(levelPrefixArray[level]+dataFilePrefix+suffix, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	buf := bufio.NewWriterSize(file, defaultBufferSize)

	kIndexes := make([]*util.KIndex, 0)

	start, end := 0, 0

	for index, entity := range kv {
		entityBytes := entity.ToByteArray()
		_, err := buf.Write(entityBytes)
		if err != nil {
			log.Fatalln(err)
			return err
		}
		if closingBound < buf.Buffered() {
			err = buf.Flush()
			if err != nil {
				return err
			}
		}
		end += len(entityBytes)
		if index%indexRangeSize == 0 {
			kIndexes = append(kIndexes, util.NewKIndex(entity.Key, start, end))
			start = end + 1
		}
	}
	if kIndexes[len(kIndexes)-1].GetKey() != kv[len(kv)-1].Key {
		kIndexes = append(kIndexes, util.NewKIndex(kv[len(kv)-1].Key, start, end))
	}
	err = buf.Flush()
	if err != nil {
		return err
	}
	err = writeDBIndexToFile(suffix, kIndexes, level)
	if err != nil {
		return err

	}
	return nil
}

func writeLevel0DBIndexToFile(suffix string, koffset []*util.KIndex) error {
	return writeDBIndexToFile(suffix, koffset, 0)
}

func writeDBIndexToFile(suffix string, koffset []*util.KIndex, level int) error {
	file, err := os.OpenFile(levelPrefixArray[level]+indexFilePrefix+suffix, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	buf := bufio.NewWriterSize(file, defaultBufferSize)
	for _, entity := range koffset {
		_, err := buf.Write(entity.ToByteArray())
		if err != nil {
			log.Fatalln(err)
			return err
		}
		if closingBound < buf.Buffered() {
			err = buf.Flush()
			if err != nil {
				return err
			}
		}
	}
	err = buf.Flush()
	if err != nil {
		return err
	}
	return nil
}

func readAllDBDataFromFile(suffix string) ([]*util.KV, error) {
	kvArray := make([]*util.KV, 0)
	file, err := os.OpenFile(level0FilePathPrefix+dataFilePrefix+suffix, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	buf := bufio.NewReaderSize(file, defaultBufferSize)
	for {
		keyBytes, err := buf.ReadString(sep)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		valueBytes, err := buf.ReadString(sep)
		if err != nil {
			return nil, err
		}
		key, err := strconv.Atoi(keyBytes[:len(keyBytes)-1])
		if err != nil {
			return nil, err
		}
		kvArray = append(kvArray, util.NewKV(key, []byte(valueBytes[:len(valueBytes)-1])))
	}

	return kvArray, nil
}

func readRangeDBDataFromFile(suffix string, start int, end int) ([]*util.KV, error) {
	kvArray := make([]*util.KV, 0)
	file, err := os.OpenFile(level0FilePathPrefix+dataFilePrefix+suffix, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)
	buf := bufio.NewReaderSize(file, defaultBufferSize)
	_, err = buf.Discard(start - 1)
	if err != nil {
		return nil, err
	}
	size := end - start
	readBytesSize := 0
	for readBytesSize < size {
		keyBytes, err := buf.ReadString(sep)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		valueBytes, err := buf.ReadString(sep)
		if err != nil {
			return nil, err
		}
		key, err := strconv.Atoi(keyBytes[:len(keyBytes)-1])
		if err != nil {
			return nil, err
		}
		kvArray = append(kvArray, util.NewKV(key, []byte(valueBytes[:len(valueBytes)-1])))
		readBytesSize += len(keyBytes) + len(valueBytes)
	}

	return kvArray, nil
}

// searchKeyFromFile
// return :
// []byte => the value mapped by special key
// bool   => if key exist and program has no error,
//			 this return value will be true
//           otherwise be false
// error  => error
func searchKeyFromFile(key int) ([]byte, bool, error) {
	suffixArray, err := getFileSuffixList()
	if err != nil {
		return nil, false, err
	}
	for _, suffix := range suffixArray {
		kIndexArray, err := readDBIndexFromFile(suffix)
		if err != nil {
			return nil, false, err
		}
		left := 0
		right := len(kIndexArray) - 1

		if kIndexArray[left].GetKey() == key {
			l := kIndexArray[left].GetStart()
			if l == 0 {
				l = 1
			}
			r := kIndexArray[left].GetEnd()
			kvArray, err := readRangeDBDataFromFile(suffix, l, r)
			if err != nil {
				return nil, false, err
			}
			res, ok := searchKeyFromKVArray(key, kvArray)
			if ok {
				return res.Value, true, nil
			} else {
				continue
			}
		}

		if kIndexArray[right].GetKey() == key {
			fmt.Println(kIndexArray[right])
			kvArray, err := readRangeDBDataFromFile(suffix, kIndexArray[right].GetStart(), kIndexArray[right].GetEnd())
			if err != nil {
				return nil, false, err
			}
			res, ok := searchKeyFromKVArray(key, kvArray)
			if ok {
				return res.Value, true, nil
			} else {
				continue
			}
		}

		if kIndexArray[left].GetKey() > key || kIndexArray[right].GetKey() < key {
			// key value don't include in this file
			// because key bigger than max value or smaller than min value in this file
			//Can't find in this file,next
			continue
		}

		for right-left > 1 {
			mid := (left + right) / 2
			if kIndexArray[mid].GetKey() >= key {
				right = mid
			} else {
				left = mid
			}
		}
		kIndex := kIndexArray[right]
		kvArray, err := readRangeDBDataFromFile(suffix, kIndex.GetStart(), kIndex.GetEnd())
		if err != nil {
			return nil, false, err
		}
		res, ok := searchKeyFromKVArray(key, kvArray)
		if ok {
			return res.Value, true, nil
		}
	}

	return nil, false, nil
}

// true =>  exist
// false => unexist
func searchKeyFromKVArray(key int, kvArray []*util.KV) (*util.KV, bool) {
	left := 0
	right := len(kvArray) - 1
	for right-left > 1 {
		mid := (left + right) / 2
		if key < kvArray[mid].Key {
			right = mid
		} else if key > kvArray[mid].Key {
			left = mid
		} else {
			return kvArray[mid], true
		}
	}
	if key == kvArray[left].Key {
		return kvArray[left], true
	}
	if key == kvArray[right].Key {
		return kvArray[right], true
	}
	return nil, false
}

func getFileSuffixList() ([]string, error) {
	res := make([]string, 0)
	fileNameList, err := getLevel0DBDataFileNameList()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(fileNameList); i++ {
		tempArray := strings.Split(fileNameList[i], "_")
		res = append(res, tempArray[len(tempArray)-1])
	}
	return res, nil
}

func readDBIndexFromFile(suffix string) ([]*util.KIndex, error) {
	kIndexArray := make([]*util.KIndex, 0)
	file, err := os.OpenFile(level0FilePathPrefix+indexFilePrefix+suffix, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(file)

	buf := bufio.NewReaderSize(file, defaultBufferSize)
	for {
		keyBytes, err := buf.ReadString(sep)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		startBytes, err := buf.ReadString(sep)
		if err != nil {
			return nil, err
		}
		endBytes, err := buf.ReadString(sep)
		if err != nil {
			return nil, err
		}
		key, err := strconv.Atoi(keyBytes[:len(keyBytes)-1])
		if err != nil {
			return nil, err
		}
		start, err := strconv.Atoi(startBytes[:len(startBytes)-1])
		if err != nil {
			return nil, err
		}
		end, err := strconv.Atoi(endBytes[:len(endBytes)-1])
		if err != nil {
			return nil, err
		}
		kIndexArray = append(kIndexArray, util.NewKIndex(key, start, end))
	}

	return kIndexArray, nil
}

func freezeDataToFile(list *skiplist.SkipList) error {
	suffixString, err := nextDBFileSuffix()
	if err != nil {
		return err
	}
	oldKV, err := readAllDBDataFromFile(suffixString)
	if err != nil {
		return err
	}
	newKV := list.ExportAllElement()
	res := KVArrayMerge(oldKV, newKV)
	err = writeLevel0DBToFile(suffixString, res)
	if err != nil {
		return err
	}
	return nil
}

func nextDBFileSuffix() (string, error) {
	dir, err := os.ReadDir(currentStoragePath)
	if err != nil {
		return "", nil
	}

	var res os.DirEntry = nil

	for _, entity := range dir {
		if !entity.IsDir() && strings.HasPrefix(entity.Name(), dataFilePrefix) {
			res = entity
		}
	}
	timeString := strconv.FormatInt(time.Now().Unix(), 10)
	if res == nil {
		return timeString, nil
	}
	info, err := res.Info()
	if info.Size() > level1MaxSize {
		return timeString, nil
	}
	splitArray := strings.Split(info.Name(), "_")
	fileTimeString := splitArray[len(splitArray)-1]
	return fileTimeString, nil

}

func getLevel0DBDataFileNameList() ([]string, error) {
	dir, err := os.ReadDir(currentStoragePath)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for i := len(dir) - 1; i >= 0; i-- {
		if !dir[i].IsDir() && strings.HasPrefix(dir[i].Name(), dataFilePrefix) {
			res = append(res, dir[i].Name())
		}
	}
	return res, nil
}

func getLevel0DBIndexFileNameList() ([]string, error) {
	dir, err := os.ReadDir(currentStoragePath)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for i := len(dir) - 1; i >= 0; i-- {
		if !dir[i].IsDir() && strings.HasPrefix(dir[i].Name(), indexFilePrefix) {
			res = append(res, dir[i].Name())
		}
	}
	return res, nil
}

func MergeLevel0Directory() error {
	list, err := getFileSuffixList()
	kvs := make([]*util.KV, 0)
	if err != nil {
		return err
	}
	for _, suffix := range list {
		tempKVArrays, err := readAllDBDataFromFile(suffix)
		if err != nil {
			return err
		}
		kvs = KVArrayMerge(kvs, tempKVArrays)
	}
	err = writeDBToFile(strconv.FormatInt(time.Now().Unix(), 10), kvs, 1)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	dirArray := []string{
		storageFilePathPrefix,
		level0FilePathPrefix,
		level1FilePathPrefix,
		level2FilePathPrefix,
		level3FilePathPrefix,
		level4FilePathPrefix,
		level5FilePathPrefix,
	}
	for _, dir := range dirArray {
		createDirectoryIfNotExist(dir)
	}
}

func createDirectoryIfNotExist(dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			log.Fatalln(err)
			return
		}
	}
}
