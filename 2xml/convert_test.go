package cvt2xml

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-xmlfmt/xmlfmt"
	errs "github.com/nsip/sif-json2xml/err-const"
)

func TestJSONRoot(t *testing.T) {
	bytes, err := ioutil.ReadFile("../data/examples/3.4.6/Activity.json")
	failOnErr("%v", err)
	fPln(jsonRoot(string(bytes)))
}

func j2x(dim int, tid int, done chan int, params ...interface{}) {
	defer func() { done <- tid }()
	ver := params[0].(string)
	files := params[1].([]os.FileInfo)
	dir := params[2].(string)

	for i := tid; i < len(files); i += dim {
		ResetAll()

		obj := rmTailFromLast(files[i].Name(), ".")
		bytes, err := ioutil.ReadFile(filepath.Join(dir, files[i].Name()))
		failOnErr("%v", err)

		sif, sv, err := JSON2XML(string(bytes), ver)
		failOnErr("%v", err)

		sif = xmlfmt.FormatXML(sif, "", "    ")
		sif = sTrim(sif, " \t\n\r")

		fPf("%-40s%-10sapplied\n", obj, sv)
		if sif != "" {
			mustWriteFile(fSf("../data/output/%s/%s.xml", sv, obj), []byte(sif))
		}
	}
}

func TestJSON2XML(t *testing.T) {
	defer trackTime(time.Now())
	// enableLog2F(true, "./error.log")
	// defer enableLog2F(false, "")

	ver := "3.4.8"
	dir := `../data/examples/` + ver
	// dir := `../data/examples/temp/`
	files, err := ioutil.ReadDir(dir)
	failOnErr("%v", err)
	failOnErrWhen(len(files) == 0, "%v", errs.FILE_NOT_FOUND)
	syncParallel(1, j2x, ver, files, dir) // only dispatch 1 goroutine, otherwise, error
	fPln("OK")
}
