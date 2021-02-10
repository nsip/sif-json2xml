package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf16"

	"github.com/cdutwhu/debog/fn"
	"github.com/cdutwhu/gotil/io"
	"github.com/cdutwhu/gotil/judge"
	"github.com/cdutwhu/gotil/net"
	"github.com/cdutwhu/gotil/rflx"
	"github.com/cdutwhu/n3-util/n3log"
	"github.com/cdutwhu/n3-util/rest"
)

var (
	fPln             = fmt.Println
	fSf              = fmt.Sprintf
	fPf              = fmt.Printf
	fEf              = fmt.Errorf
	sReplaceAll      = strings.ReplaceAll
	sTrimRight       = strings.TrimRight
	sTrimLeft        = strings.TrimLeft
	sTrim            = strings.Trim
	sIndex           = strings.Index
	sLastIndex       = strings.LastIndex
	sJoin            = strings.Join
	sNewReader       = strings.NewReader
	sHasSuffix       = strings.HasSuffix
	sHasPrefix       = strings.HasPrefix
	sContains        = strings.Contains
	sSplit           = strings.Split
	rxMustCompile    = regexp.MustCompile
	failOnErr        = fn.FailOnErr
	failOnErrWhen    = fn.FailOnErrWhen
	enableLog2F      = fn.EnableLog2F
	enableWarnDetail = fn.EnableWarnDetail
	logWhen          = fn.LoggerWhen
	logger           = fn.Logger
	warnOnErr        = fn.WarnOnErr
	warner           = fn.Warner
	localIP          = net.LocalIP
	env2Struct       = rflx.Env2Struct
	struct2Map       = rflx.Struct2Map
	tryInvoke        = rflx.TryInvoke
	loggly           = n3log.Loggly
	logBind          = n3log.Bind
	setLoggly        = n3log.SetLoggly
	syncBindLog      = n3log.SyncBindLog
	isXML            = judge.IsXML
	isJSON           = judge.IsJSON
	exist            = judge.Exist
	mustWriteFile    = io.MustWriteFile
	url1Value        = rest.URL1Value
)

var (
	logGrp     = logBind(logger) // logBind(logger, loggly("info"))
	warnGrp    = logBind(warner) // logBind(warner, loggly("warn"))
	rxRootAttr = rxMustCompile(`^\{[\n\s]*"@\w+":\s+"[^"]*"`)
)

func initMutex(route interface{}) map[string]*sync.Mutex {
	mMtx := make(map[string]*sync.Mutex)
	for _, v := range struct2Map(route) {
		mMtx[v.(string)] = &sync.Mutex{}
	}
	return mMtx
}

// UTF16To8 :
func UTF16To8(s string, order binary.ByteOrder) string {
	if len(s) == 0 {
		return ""
	}
	b := []byte(s)
	switch b[0] {
	case 0xff, 0xfe:
		b = b[2:]
		break
	default:
		return s
	}

	ints := make([]uint16, len(b)/2)
	if err := binary.Read(bytes.NewReader(b), order, &ints); err != nil {
		panic(err)
	}
	return string(utf16.Decode(ints))
}

func aln2mav(aln, apf string) map[string]interface{} {
	mav := make(map[string]interface{})
	if apf == "" {
		apf = "@"
	}
	aln = sTrim(sTrimLeft(aln, "{"), " \t\n\r")
	av := sSplit(aln, "\":")
	if len(av) != 2 {
		return nil
	}
	av[0] = sTrimLeft(av[0], "\"")
	av[0] = sTrimLeft(av[0], apf)
	av[1] = sTrimLeft(av[1], " ")
	av[1] = sTrim(av[1], "\"")
	mav[av[0]] = av[1]
	return mav
}
