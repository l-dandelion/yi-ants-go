package module

import (
	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

//constant of module type
const (
	TYPE_DOWNLOADER int8 = 0
	TYPE_ANALYZER   int8 = 1
	TYPE_PIPELINE   int8 = 2

	LETTER_DOWNLOADER = "D"
	LETTER_ANALYZER   = "A"
	LETTER_PIPELINE   = "P"

	TYPE_DOWNLOADER_DESC = "Downloader(下载器)"
	TYPE_ANALYZER_DESC   = "Analyzer(解析器)"
	TYPE_PIPELINE_DESC   = "Pipeline(条目处理器)"
)

//type-letter map
var legalTypeLetterMap = map[int8]string{
	TYPE_DOWNLOADER: LETTER_DOWNLOADER,
	TYPE_ANALYZER:   LETTER_ANALYZER,
	TYPE_PIPELINE:   LETTER_PIPELINE,
}

//letter-type map
var legalLetterTypeMap = map[string]int8{
	LETTER_DOWNLOADER: TYPE_DOWNLOADER,
	LETTER_ANALYZER:   TYPE_ANALYZER,
	LETTER_PIPELINE:   TYPE_PIPELINE,
}

//type-description map
var typeDescMap = map[int8]string{
	TYPE_DOWNLOADER: TYPE_DOWNLOADER_DESC,
	TYPE_ANALYZER:   TYPE_ANALYZER_DESC,
	TYPE_PIPELINE:   TYPE_PIPELINE_DESC,
}

/*
 * check whether the type of module instance is matched
 */
func IsMatch(mtype int8, module Module) bool {
	if module == nil {
		return false
	}

	switch mtype {
	case TYPE_DOWNLOADER:
		if _, ok := module.(Downloader); ok {
			return true
		}
	case TYPE_ANALYZER:
		if _, ok := module.(Analyzer); ok {
			return true
		}
	case TYPE_PIPELINE:
		if _, ok := module.(Pipeline); ok {
			return true
		}
	}
	return false
}

/*
 * check whether the type is legal
 */
func LegalType(mtype int8) bool {
	_, ok := legalTypeLetterMap[mtype]
	return ok
}

/*
 * check whether the type letter if legal
 */
func LegalLetter(letter string) bool {
	_, ok := legalLetterTypeMap[letter]
	return ok
}

/*
 * get module type from MID
 */
func GetType(mid MID) (mtype int8, err *constant.YiError) {
	part, err := SplitMID(mid)
	if err != nil {
		return
	}
	mtype, _ = letter2Type(part[0])
	return
}

/*
 * get type letter by type
 */
func type2Letter(mtype int8) (letter string, ok bool) {
	letter, ok = legalTypeLetterMap[mtype]
	return
}

/*
 * get type by type letter
 */
func letter2Type(letter string) (mtype int8, ok bool) {
	mtype, ok = legalLetterTypeMap[letter]
	return
}
