package cvt2xml

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cdutwhu/debog/fn"
	"github.com/cdutwhu/gotil/dispatcher"
	"github.com/cdutwhu/gotil/endec"
	"github.com/cdutwhu/gotil/io"
	"github.com/cdutwhu/gotil/judge"
	"github.com/cdutwhu/gotil/misc"
	"github.com/cdutwhu/gotil/net"
	"github.com/cdutwhu/gotil/rflx"
	"github.com/cdutwhu/gotil/str"
	jt "github.com/cdutwhu/json-tool"
	xt "github.com/cdutwhu/xml-tool"
)

var (
	fPf                = fmt.Printf
	fPln               = fmt.Println
	fPt                = fmt.Print
	fSp                = fmt.Sprint
	fSf                = fmt.Sprintf
	sHasPrefix         = strings.HasPrefix
	sHasSuffix         = strings.HasSuffix
	sReplaceAll        = strings.ReplaceAll
	failOnErr          = fn.FailOnErr
	failOnErrWhen      = fn.FailOnErrWhen
	failP1OnErrWhen    = fn.FailP1OnErrWhen
	failP1OnErr        = fn.FailP1OnErr
	enableLog2F        = fn.EnableLog2F
	warner             = fn.Warner
	localIP            = net.LocalIP
	sSplit             = strings.Split
	sReplace           = strings.Replace
	sCount             = strings.Count
	sTrim              = strings.Trim
	sTrimLeft          = strings.TrimLeft
	sIndex             = strings.Index
	sContains          = strings.Contains
	sSpl               = strings.Split
	sJoin              = strings.Join
	sNewReader         = strings.NewReader
	rxMustCompile      = regexp.MustCompile
	splitRev           = str.SplitRev
	rmTailFromLast     = str.RmTailFromLast
	rmHeadToLast       = str.RmHeadToLast
	rmTailFromFirstAny = str.RmTailFromFirstAny
	replByPosGrp       = str.ReplByPosGrp
	indent             = str.IndentTxt
	syncParallel       = dispatcher.SyncParallel
	trackTime          = misc.TrackTime
	isJSON             = judge.IsJSON
	md5Str             = endec.MD5Str
	mapsMerge          = rflx.MapMerge
	mustWriteFile      = io.MustWriteFile
	xmlRoot            = xt.Root
	jsonRoot           = jt.Root
)

const MaxGoTo = 100

var (
	nGoTo = 0

	rx1       = rxMustCompile(`\n\s*<#content>`)
	rx2       = rxMustCompile(`</#content>\n\s*`)
	rx3       = rxMustCompile(`".+": ".*(\\n)+.*"`)
	rxTag     = rxMustCompile(`<.+[> ]`)
	rxAttr    = rxMustCompile(`.+="`)
	rxReplNum = rxMustCompile(`@\d+#`)

	TrvsGrpViaSpec []string                    // from SIF Spec
	mPathAttrs     = make(map[string][]string) // key: spec path, value: attribute-value
	mPathAttrIdx   = make(map[string]int)      // key: spec path, value: attribute-index

	mObjIdxStart  = make(map[string]int)    // key: obj-type@level, value: line-number
	mObjIdxEnd    = make(map[string]int)    // key: obj-type@level, value: line-number
	mOAPrtLn      = make(map[string]int)    // key: obj
	mIPathSubXML  = make(map[string]string) // key: path@index
	mIPathSubMark = make(map[string]string) // key: path@index
	mPathIdx      = make(map[string]int)    // key: path, for IPath

	RewindAttrIter = func() {
		for k := range mPathAttrIdx {
			mPathAttrIdx[k] = 0
		}
	}

	ResetPrt = func() {
		mOAPrtLn = make(map[string]int)
	}

	ResetAll = func() {
		mObjIdxStart = make(map[string]int)
		mObjIdxEnd = make(map[string]int)
		mOAPrtLn = make(map[string]int)
		mIPathSubXML = make(map[string]string)
		mIPathSubMark = make(map[string]string)
		mPathIdx = make(map[string]int)
		nGoTo = 0
	}

	mkIndent = func(n int) (indent string) {
		const INDENT = "    " // 4 space
		for i := 0; i < n; i++ {
			indent += INDENT
		}
		return
	}
)

var (
	mOldNew = map[string]string{
		" lang=\"": " xml:lang=\"",
	}
)
