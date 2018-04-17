package module

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

// the template of ID of module
var midTemplate = "%s%d|%s"

// the ID of module
type MID string

/*
 * generate MID according to type, sn and netword address
 */
func GenMID(mtype int8, sn uint64, maddr net.Addr) (MID, *constant.YiError) {
	if !LegalType(mtype) {
		return "", constant.NewYiErrorf(constant.ERR_GENERATE_MID, "Illegal module type: %s", mtype)
	}
	letter := legalTypeLetterMap[mtype]
	var midStr string
	if maddr == nil {
		midStr = fmt.Sprintf(midTemplate, letter, sn, "")
		midStr = midStr[:len(midStr)-1]
	} else {
		midStr = fmt.Sprintf(midTemplate, letter, sn, maddr.String())
	}
	return MID(midStr), nil
}

/*
 * check whether the mid is vaild
 */
func LegalMID(mid MID) bool {
	if _, err := SplitMID(mid); err == nil {
		return true
	}
	return false
}

/*
 * split MID
 * if success, return three strings(the letter of type, sn and netword address)
 * else return an error
 */
func SplitMID(mid MID) ([]string, *constant.YiError) {
	var letter, snStr, addr string

	//check the len of mid
	midStr := string(mid)
	if len(midStr) <= 1 {
		return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
			"Illegal MID string(the len of midStr less than 2): %s", midStr)
	}

	//type letter
	letter = midStr[:1]
	if !LegalLetter(letter) {
		return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
			"Illegal module type letter: %s", letter)
	}

	//split SN and network address
	snAndAddr := midStr[1:]
	index := strings.LastIndex(snAndAddr, "|")
	if index < 0 {
		snStr = snAndAddr
	} else {
		snStr = snAndAddr[:index]
		addr = snAndAddr[index+1:]

		//split ip and port and check
		index = strings.LastIndex(addr, ":")
		if index <= 0 {
			return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
				"Illegal module address: %s", addr)
		}

		ipStr := addr[:index]
		if ip := net.ParseIP(ipStr); ip == nil {
			return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
				"Illegal module IP: %s", ipStr)
		}

		portStr := addr[index+1:]
		if _, err := strconv.ParseUint(portStr, 10, 64); err != nil {
			return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
				"Illegal module port: %s", portStr)
		}
	}

	//check SN
	if !legalSN(snStr) {
		return nil, constant.NewYiErrorf(constant.ERR_SPLIT_MID,
			"Illegal module SN: %s", snStr)
	}

	return []string{letter, snStr, addr}, nil
}

/*
 * check whether SN is valid.
 */
func legalSN(snStr string) bool {
	_, err := strconv.ParseUint(snStr, 10, 64)
	return err == nil
}
