package cvt2xml

import (
	"encoding/json"
	"fmt"

	"github.com/clbanning/mxj"
	cfg "github.com/nsip/sif-json2xml/config/cfg"
	errs "github.com/nsip/sif-json2xml/err-const"
	sif346 "github.com/nsip/sif-json2xml/sif-spec/3.4.6"
	sif347 "github.com/nsip/sif-json2xml/sif-spec/3.4.7"
)

// ----------------------------------------- //

// NextAttr : From Spec
func NextAttr(obj, fullpath string) (value string, end bool) {
	if fullpath == "/" {
		fullpath = obj
	} else {
		fullpath += obj
	}
	idx := mPathAttrIdx[fullpath]
	if idx == len(mPathAttrs[fullpath]) {
		return "", true
	}
	defer func() {
		mPathAttrIdx[fullpath]++
	}()
	return mPathAttrs[fullpath][idx], false
}

// PrintXML : append print to a string
func PrintXML(paper, line, mark string, iLine int, tag string) (string, bool) {
	if _, ok := mOAPrtLn[tag]; !ok {
		mOAPrtLn[tag] = -1
	}

	if iLine <= mOAPrtLn[tag] {
		return paper, false
	}
	mOAPrtLn[tag] = iLine

	if mark != "" {
		return paper + line + mark + "\n", true
	}
	return paper + line + "\n", true
}

// SortSimpleObject : xml is 4 space formatted, level is obj level
// obj [level] = attribute [level-1]
// NextAttr is available
func SortSimpleObject(xml, obj string, level int, trvsPath string) (paper string) {
	defer func() {
		ResetPrt()
	}()

	const INDENT = "    " // 4 space
	indentObj, indentAttr := "", INDENT
	for i := 0; i < level; i++ {
		indentObj += INDENT
		indentAttr += INDENT
	}

	OS1 := fSf("%s<%s ", indentObj, obj)  // begin 1
	OS2 := fSf("%s<%s>", indentObj, obj)  // begin 2
	OS3 := fSf("%s<%s/>", indentObj, obj) // empty element begin & end
	OE1 := fSf("%s</%s>", indentObj, obj) // block end

	lines := sSplit(xml, "\n")

	// Find nObj
	nObj := sCount(xml, OS1)
	if n := sCount(xml, OS2); n > nObj {
		nObj = n
	}

	objIdx := fSf("%s@%d", obj, level)
	if _, ok := mObjIdxStart[objIdx]; !ok {
		mObjIdxStart[objIdx] = -1
	}
	if _, ok := mObjIdxEnd[objIdx]; !ok {
		mObjIdxEnd[objIdx] = -1
	}

	RewindAttrIter()
	PS, PE := -1, -1

	// ---------------------------------- //
	for i, l := range lines {
		if (sHasPrefix(l, OS1) || sHasPrefix(l, OS2) || sHasPrefix(l, OS3)) && i > mObjIdxStart[objIdx] {
			if _, ok := PrintXML(paper, l, "", i, "*"+obj); ok { // [*+obj] is probe to detect Start line
				PS, mObjIdxStart[objIdx] = i, i
			}
		}
		if sHasPrefix(l, OE1) && i > mObjIdxEnd[objIdx] {
			if _, ok := PrintXML(paper, l, "", i, "*/"+obj); ok { // [*/+obj] is probe to detect End line
				PE, mObjIdxEnd[objIdx] = i, i
			}
		}
		if PS != -1 && PE != -1 { // if all found, break. PE may not be found when Single Line Attribute
			break
		}
	}
	// ---------------------------------- //

	paper, _ = PrintXML(paper, lines[PS], fSf("@%d#", PS), PS, obj)

	attr, end := NextAttr(obj, trvsPath)
	for ; !end; attr, end = NextAttr(obj, trvsPath) {
		AS1 := fSf("%s<%s ", indentAttr, attr)  // begin 1
		AS2 := fSf("%s<%s>", indentAttr, attr)  // begin 2
		AS3 := fSf("%s<%s/>", indentAttr, attr) // empty element begin & end
		AE1 := fSf("%s</%s>", indentAttr, attr) // block end
		AE2 := fSf("</%s>", attr)               // one line end
		for i, l := range lines {
			if i > PS && i < PE {
				switch {
				case ((sHasPrefix(l, AS1) || sHasPrefix(l, AS2)) && sHasSuffix(l, AE2)) || sHasPrefix(l, AS3): // one line (including empty)
					if tempPaper, ok := PrintXML(paper, l, "", i, attr); ok {
						paper = tempPaper
					}
				case sHasPrefix(l, AS1) || sHasPrefix(l, AS2): // sub-object START
					if tempPaper, ok := PrintXML(paper, l, fSf("@%d#...", i), i, attr); ok {
						paper = tempPaper
					}
				case sHasPrefix(l, AE1): // sub-object END
					if tempPaper, ok := PrintXML(paper, l, "", i, "/"+attr); ok {
						paper = tempPaper
					}
				}
			}
		}
	}

	// Single Line Object has NO End Tag
	if PE != -1 {
		paper, _ = PrintXML(paper, lines[PE], "", PE, "/"+obj)
	}

	return
}

// From : AGAddressCollectionSubmission~AddressCollectionReportingList@0~AddressCollectionReporting@0~EntityContact@0
// To :   AGAddressCollectionSubmission/AddressCollectionReportingList/AddressCollectionReporting/EntityContact/
func iPath2SpecPath(iPath, oldSep, newSep string) string {
	ss := sSpl(iPath, oldSep)
	for i, s := range ss {
		ss[i] = rmTailFromLast(s, "@")
	}
	return sJoin(ss, newSep) + newSep
}

// ExtractOA : root obj, path is ""
func ExtractOA(xml, obj, path string, lvl int) string {
	const (
		iPathSep, specTrvsPathSep = "~", "/"
	)

	S := mkIndent(lvl+1) + "<"
	E := S + "/"

	// fPln(path)
	// fPln(iPath2SpecPath(path, iPathSep, specTrvsPathSep))
	specTrvsPath := iPath2SpecPath(path, iPathSep, specTrvsPathSep)

	lvlOAs := []string{} // Complex Object Tags
	xmlobj := sTrim(SortSimpleObject(xml, obj, lvl, specTrvsPath), "\n")
	for _, l := range sSplit(xmlobj, "\n") {
		sl := 0
		switch {
		case sHasPrefix(l, S) && !sHasPrefix(l, E) && sContains(l, "..."): // Complex Object Tags
			sl = len(S)
		default:
			continue
		}
		oa := rmTailFromFirstAny(l[sl:], " ", ">")
		lvlOAs = append(lvlOAs, oa)
	}

	if path != "" {
		path += (iPathSep + obj)
	} else {
		path = obj
	}
	if _, ok := mPathIdx[path]; !ok {
		mPathIdx[path] = 0
	}

	iPath := fSf("%s@%d", path, mPathIdx[path])
	mPathIdx[path]++
	if path == obj { // root is without @index
		iPath = obj
	}

	mIPathSubXML[iPath] = xmlobj

	xmlobjLn1 := sSplit(xmlobj, "\n")[0]
	preBlank := mkIndent(sCount(iPath, iPathSep))
	mIPathSubMark[iPath] = fSf("%s...\n%s</%s>", xmlobjLn1, preBlank, obj)

	for _, subobj := range lvlOAs {
		ExtractOA(xml, subobj, iPath, lvl+1)
	}

	return xmlobj
}

// ----------------------------------------------- //

// JSON2XML4LF : if JSON fields have special (LF, TBL), pick them up for later replacement
func JSON2XML4LF(json string) (string, map[string]string) {
	failOnErrWhen(!isJSON(json), "%v", errs.PARAM_INVALID_JSON)
	mCodeStr := make(map[string]string)
	strGrpWithLF := rx3.FindAllString(json, -1)
	for _, s := range strGrpWithLF {
		vLiteral := sSpl(s, `": "`)[1]
		vLiteral = vLiteral[:len(vLiteral)-1]                              // literal \n \t
		vEsc := sReplaceAll(sReplaceAll(vLiteral, `\n`, "\n"), `\t`, "\t") // escape \n \t
		k4Esc := md5Str(vEsc)
		mCodeStr[k4Esc] = vEsc
		json = sReplaceAll(json, vLiteral, k4Esc)
	}
	return json, mCodeStr
}

// JSON2XML3RD : via 3rd lib converter, return Disordered, Formatted XML
func JSON2XML3RD(jstr string) string {
	var f interface{}
	failOnErr("%v", json.Unmarshal([]byte(jstr), &f))
	b, err := mxj.AnyXmlIndent(f, "", "    ", "")
	failOnErr("%v", err)

	xstr := string(b)
	xstr = sReplaceAll(xstr, "<>", "")
	xstr = sReplaceAll(xstr, "</>", "")
	xstr = rx1.ReplaceAllString(xstr, "")
	xstr = rx2.ReplaceAllString(xstr, "")
	xstr = indent(xstr, -4, false)
	xstr = sTrim(xstr, " \t\n")
	return xstr
}

// InitOAs : fill [TrvsGrpViaSpec] & [mPathAttrs] & [mPathAttrIdx]
func InitOAs(SIFSpecStr, tblSep, pathSep string) {
	if len(mPathAttrs) > 0 {
		return
	}

	const TRAVERSE = "TRAVERSE ALL, DEPTH ALL"
	for _, line := range sSplit(SIFSpecStr, "\n") {
		switch {
		case sHasPrefix(line, TRAVERSE):
			l := sTrim(line[len(TRAVERSE):], " \t\r")
			TrvsGrpViaSpec = append(TrvsGrpViaSpec, l)
		}
	}
	for _, trvs := range TrvsGrpViaSpec {
		path := sSplit(trvs, tblSep)[0]
		key := rmTailFromLast(path, pathSep)
		value := rmHeadToLast(path, pathSep)
		mPathAttrs[key] = append(mPathAttrs[key], value)
		mPathAttrIdx[key] = 0
	}
}

// AddSIFSpec : Ordered, Some pieces are different
func AddSIFSpec(xml, SIFSpecStr string) string {
	InitOAs(SIFSpecStr, "\t", "/")
	xmlhead := xml

	// adjusting attributes order
	posGrp, pathGrp, mAttrGrp, root := SearchTagWithAttr(xml)
	for i, path := range pathGrp {
		attrs2write := ""
		for _, trvs := range TrvsGrpViaSpec {
			if sHasPrefix(trvs, path+"/@") { // from Spec format
				attr := sSplit(trvs, "\t")[2][1:] // from Spec format
				attr2write := mAttrGrp[i][attr]
				attrs2write += attr2write + " "
			}
		}
		attrs2write = sTrim(attrs2write, " ")
		// fPln(attrs2write)

		start, end := posGrp[i][0], posGrp[i][1]
		xmlLn := xmlhead[start:end]
		tag, _ := TagFromXMLLine(xmlLn)
		out := fSf("%s<%s %s>", mkIndent(CountHeadSpace(xmlLn, 4)), tag, attrs2write)
		// xml = sReplByPos(xml, start, end, out)
		xml = replByPosGrp(xml, [][]int{{start, end}}, []string{out})
	}
	// End adjusting attributes order

	// ---------------------------------- //

	// Init "mIPathSubXML"
	ExtractOA(xml, root, "", 0)

	xml = mIPathSubXML[root]
AGAIN:
	for k, subxml := range mIPathSubXML {
		mark := mIPathSubMark[k]
		xml = sReplace(xml, mark, subxml, 1)
	}
	if sContains(xml, "...") {
		// mustWriteFile(fSf("./%d.xml", nGoTo), []byte(xml))
		nGoTo++
		failOnErrWhen(nGoTo > MaxGoTo, "%v: goto AGAIN", errs.INTERNAL_DEADLOCK)
		goto AGAIN
	}

	return xml
}

// -------------------------------------------------------- //

// CountHeadSpace :
func CountHeadSpace(s string, nGrp int) int {
	for i, c := range s {
		if c == ' ' {
			continue
		}
		return i / nGrp
	}
	return 0
}

// TagFromXMLLine :
func TagFromXMLLine(line string) (tag string, mKeyAttr map[string]string) {
	line = sTrim(line, " \t\n\r")
	failOnErrWhen(line[0] != '<' || line[len(line)-1] != '>', "%v: %s", errs.PARAM_INVALID_FMT, line)
	if tag := rxTag.FindString(line); tag != "" {
		tag = tag[1 : len(tag)-1] // remove '<' '>'
		ss := sSplit(tag, " ")    // cut fields
		mKeyAttr = make(map[string]string)
		for _, attr := range ss[1:] {
			if ak := rxAttr.FindString(attr); ak != "" {
				mKeyAttr[ak[:len(ak)-2]] = attr // remove '="'
			}
		}
		return ss[0], mKeyAttr
	}
	return "", nil
}

// Hierarchy :
func Hierarchy(searchArea string, lvl int, hierarchy *[]string) {
	r := rxMustCompile(fSf(`\n[ ]{%d}<.*>`, (lvl-1)*4))
	if locGrp := r.FindAllStringIndex(searchArea, -1); locGrp != nil {
		loc := locGrp[len(locGrp)-1]
		start, end := loc[0], loc[1]
		find := searchArea[start:end]
		Hierarchy(searchArea[:start], lvl-1, hierarchy)
		tag, _ := TagFromXMLLine(find)
		*hierarchy = append(*hierarchy, tag)
	}
}

// SearchTagWithAttr : where (get line from xml), tag-path (get info from spec), attribute-map (re-order attributes, reconstruct line)
func SearchTagWithAttr(xml string) (posGrp [][2]int, pathGrp []string, mAttrGrp []map[string]string, root string) {
	root = xmlRoot(xml)
	TagOrAttr := `[^ \t<>]+`
	r := rxMustCompile(fSf(`[ ]*<%[1]s[ ]+(%[1]s="%[1]s"[ ]*){2,}>`, TagOrAttr))
	// r := rxMustCompile(`[ ]*<\w+\s+(\w+="[^"]+"\s*){2,}>`)
	if loc := r.FindAllStringIndex(xml, -1); loc != nil {
		for _, l := range loc {
			hierarchy := &[]string{root}
			start, end := l[0], l[1]
			withAttr := xml[start:end]

			Hierarchy(xml[:start], CountHeadSpace(withAttr, 4), hierarchy)
			tag, mka := TagFromXMLLine(withAttr)
			*hierarchy = append(*hierarchy, tag)

			posGrp = append(posGrp, [2]int{start, end})
			pathGrp = append(pathGrp, sJoin(*hierarchy, "/"))
			mAttrGrp = append(mAttrGrp, mka)
		}
	}
	return
}

// -------------------------------------------------------- //

// JSON2XMLRepl : Pieces Replaced, should be almost identical to Original SIF
func JSON2XMLRepl(xml string, mRepl map[string]string) string {

	// remove @Number#
	xml = string(rxReplNum.ReplaceAll([]byte(xml), []byte("")))

	// others from cfg
	for old, new := range mRepl {
		xml = sReplaceAll(xml, old, new)
	}

	return xml
}

// -------------------------------------------------------- //

// JSON2XML : JSON2XML4LF -> JSON2XML3RD -> AddSIFSpec -> JSON2XMLRepl
func JSON2XML(json, sifver string) (sif, sv string, err error) {
	cfgAll := cfg.NewCfg("Config", nil, "./config/config.toml", "../config/config.toml").(*cfg.Config)

	ver := cfgAll.SIF.DefaultVer
	if sifver != "" {
		ver = sifver
	}

	var bytes []byte
	switch ver {
	case "3.4.6":
		bytes = sif346.TXT["346"]
	case "3.4.7":
		bytes = sif347.TXT["347"]
	default:
		err = fmt.Errorf("Error: No SIF Spec @ Version [%s]", ver)
		warner("%v", err)
		return
	}

	ResetAll()
	jsonWithCode, mCodeStr := JSON2XML4LF(json)
	mRepl := mapsMerge(mOldNew, mCodeStr).(map[string]string)
	return JSON2XMLRepl(AddSIFSpec(JSON2XML3RD(jsonWithCode), string(bytes)), mRepl), ver, nil
}
