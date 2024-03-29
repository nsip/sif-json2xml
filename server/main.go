package main

import (
	"context"
	"encoding/binary"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/cdutwhu/gonfig"
	"github.com/cdutwhu/gonfig/attrim"
	"github.com/cdutwhu/gonfig/strugen"
	"github.com/cdutwhu/gotil/misc"
	jt "github.com/cdutwhu/json-tool"
	"github.com/digisan/gotk/slice/ts"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	cvt "github.com/nsip/sif-json2xml/2xml"
	cfg "github.com/nsip/sif-json2xml/config/cfg"
	errs "github.com/nsip/sif-json2xml/err-const"
	sr "github.com/nsip/sif-spec-res"
)

var allSIF = []string{
	"3.4.2",
	"3.4.3",
	"3.4.4",
	"3.4.5",
	"3.4.6",
	"3.4.7",
	"3.4.8",
}

func mkCfg4Clt(cfg interface{}) {
	forel := "./config_rel.toml"
	gonfig.Save(forel, cfg)
	outToml := "./client/config.toml"
	outSrc := "./client/config.go"
	os.Remove(outToml)
	os.Remove(outSrc)
	attrim.SelCfgAttrL1(forel, outToml, "Service", "Route", "Server", "Access")
	strugen.GenStruct(outToml, "Config", "client", outSrc)
	strugen.GenNewCfg(outSrc)
}

func mkCfg4Docker(cfg interface{}) {
	forel := "./config_rel.toml"
	gonfig.Save(forel, cfg)
	outToml := "../config_d.toml"
	os.Remove(outToml)
	attrim.RmCfgAttrL1(forel, outToml, "Log", "Server", "Access")
}

var (
	gCfg *cfg.Config
)

func main() {
	// Load global config.toml file from Config/
	gonfig.SetDftCfgVal("sif-json2xml", "0.0.0")
	pCfg := cfg.NewCfg(
		"Config",
		map[string]string{
			"[s]":    "Service",
			"[v]":    "Version",
			"[port]": "WebService.Port",
		},
		"./config/config.toml",
		"../config/config.toml",
	)
	failOnErrWhen(pCfg == nil, "%v: Config Init Error", errs.CFG_INIT_ERR)
	gCfg = pCfg.(*cfg.Config)

	// Trim a shorter config toml file for client package
	if len(os.Args) > 2 && os.Args[2] == "trial" {
		mkCfg4Docker(gCfg)
		mkCfg4Clt(gCfg)
		return
	}

	ws := gCfg.WebService
	var IService interface{} = gCfg.Service // Cfg.Service can be "string" or "interface{}"
	service := IService.(string)

	// Set Jaeger Env for tracing
	os.Setenv("JAEGER_SERVICE_NAME", service)
	os.Setenv("JAEGER_SAMPLER_TYPE", "const")
	os.Setenv("JAEGER_SAMPLER_PARAM", "1")

	// Set LOGGLY
	setLoggly(false, gCfg.Loggly.Token, service)

	// Set Log Options
	syncBindLog(true)
	enableWarnDetail(false)
	if gCfg.Log != "" {
		enableLog2F(true, gCfg.Log)
		logGrp.Do(fSf("local log file @ [%s]", gCfg.Log))
	}

	logGrp.Do(fSf("[%s] Hosting on: [%v:%d], version [%v]", service, localIP(), ws.Port, gCfg.Version))

	// Start Service
	done := make(chan string)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt)
	go HostHTTPAsync(c, done)
	logGrp.Do(<-done)
}

func shutdownAsync(e *echo.Echo, sig <-chan os.Signal, done chan<- string) {
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	failOnErr("%v", e.Shutdown(ctx))
	time.Sleep(20 * time.Millisecond)
	done <- "Shutdown Successfully"
}

// HostHTTPAsync : Host a HTTP Server for SIF or JSON
func HostHTTPAsync(sig <-chan os.Signal, done chan<- string) {
	defer logGrp.Do("HostHTTPAsync Exit")

	e := echo.New()
	defer e.Close()

	// waiting for shutdown
	go shutdownAsync(e, sig, done)

	// Add Jaeger Tracer into Middleware
	c := jaegertracing.New(e, nil)
	defer c.Close()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit("2G"))
	// CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{echo.GET, echo.POST},
		AllowCredentials: true,
	}))

	e.Logger.SetOutput(os.Stdout)
	e.Logger.Infof(" ------------------------ e.Logger.Infof ------------------------ ")

	var (
		// Cfg    = rflx.Env2Struct("Config", &cfg.Config{}).(*cfg.Config)
		port   = gCfg.WebService.Port
		fullIP = localIP() + fSf(":%d", port)
		route  = gCfg.Route
		mMtx   = initMutex(&gCfg.Route)
		vers   = sr.GetAllVer("v", "")
	)

	// prepare for inferring 'wrapped'
	mAllObj := make(map[string][]string)
	for _, v := range allSIF {
		mAllObj[v], _ = cvt.AllSIFObject(v)
	}

	defer e.Start(fSf(":%d", port))
	logGrp.Do("Echo Service is Starting ...")

	// *************************************** List all API, FILE *************************************** //

	path := route.Help
	e.GET(path, func(c echo.Context) error {
		defer mMtx[path].Unlock()
		mMtx[path].Lock()

		return c.String(http.StatusOK,
			fSf("Converter Server Version: %s\n\n", gCfg.Version)+
				fSf("API:\n\n [POST] %-40s\n%s", fullIP+route.Convert,
					"\n Description: Upload SIF(JSON), return SIF(XML)\n"+
						"\n Parameters:"+
						"\n -- [sv]:   available SIF Ver: "+fSf("%v", vers)+
						"\n -- [nats]: send json to NATS?"+
						"\n -- [wrap]: is uploaded SIF file with single wrapped root?"))
	})

	// ------------------------------------------------------------------------------------ //

	// mRouteRes := map[string]string{
	// 	"/client-linux64": Cfg.File.ClientLinux64,
	// 	"/client-mac":     Cfg.File.ClientMac,
	// 	"/client-win64":   Cfg.File.ClientWin64,
	// 	"/client-config":  Cfg.File.ClientConfig,
	// }

	// routeFun := func(rt, res string) func(c echo.Context) error {
	// 	return func(c echo.Context) (err error) {
	// 		if _, err = os.Stat(res); err == nil {
	// 			fPln(rt, res)
	// 			return c.File(res)
	// 		}
	// 		return warnOnErr("%v: [%s]  get [%s]", errs.FILE_NOT_FOUND, rt, res)
	// 	}
	// }

	// for rt, res := range mRouteRes {
	// 	e.GET(rt, routeFun(rt, res))
	// }

	// -------------------------------------------------------------------------------- //
	// -------------------------------------------------------------------------------- //

	path = route.Convert
	e.POST(path, func(c echo.Context) error {
		defer misc.TrackTime(time.Now())
		defer mMtx[path].Unlock()
		mMtx[path].Lock()

		var (
			status  = http.StatusOK
			Ret     string
			RetSB   strings.Builder
			results []reflect.Value        // for 'jaegertracing.TraceFunction'
			mav     map[string]interface{} // for wrapper root attributes
		)

		logGrp.Do("Parsing Params")
		pvalues, sv, wrapped := c.QueryParams(), "", false
		if ok, v := url1Value(pvalues, 0, "sv"); ok {
			sv = v
		}
		if ok, w := url1Value(pvalues, 0, "wrap"); ok && w != "false" {
			wrapped = true
		}

		logGrp.Do("Reading Body")
		bytes, err := ioutil.ReadAll(c.Request().Body)
		jstr, root, cont, out4ret := "", "", "", ""
		jsonObjNames, jsonContGrp := []string{}, []string{}

		if err != nil {
			status = http.StatusInternalServerError
			RetSB.Reset()
			RetSB.WriteString(err.Error() + " @Request Body")
			goto RET
		}
		if jstr = string(bytes); len(jstr) == 0 {
			status = http.StatusBadRequest
			RetSB.Reset()
			RetSB.WriteString(errs.HTTP_REQBODY_EMPTY.Error() + " @Request Body")
			goto RET
		}
		if jstr = UTF16To8(jstr, binary.LittleEndian); !isJSON(jstr) {
			status = http.StatusBadRequest
			RetSB.Reset()
			RetSB.WriteString(errs.PARAM_INVALID_JSON.Error() + " @Request Body")
			goto RET
		}

		/// DEBUG ///
		// if sContains(jstr, "A5A575C7-8917-5101-B8E7-F08ED123A823") {
		// ioutil.WriteFile("./debug.json", []byte(jstr), 0666)
		// fPln("break")
		// }
		/// DEBUG ///

		///
		// ** if wrapped, break and handle each SIF object ** //
		///
		root, cont = jt.SglEleBlkCont(jstr) // if wrapped : => "sif", { "Activity" ... }
		// take attribute lines from cont, then
		mav = aln2mav(rxRootAttr.FindString(cont), "") // Now, Only take one wrapper root attribute

		jsonObjNames, jsonContGrp = []string{root}, []string{cont}

		// for inferring wrapped when wrap is not provided
		if !wrapped {
			wrapped = ts.NotIn(root, mAllObj[sv]...)
		}

		if wrapped {
			wrapper := jt.MkSglEleBlk(root, "~~~", false)
			out4ret = jt.Cvt2XML(wrapper, mav)
			jsonObjNames, jsonContGrp = jt.BreakMulEleBlkV2(cont) // break array to single duplicated objects
		}
		///

		for i, jsonObj := range jsonContGrp {
			obj := jsonObjNames[i]
			// logGrp.Do("cvt2json.JSON2XML")

			jsonObj = jt.MkSglEleBlk(jsonObjNames[i], jsonObj, false)

			/// DEBUG ///
			// if sContains(jsonObj, "A5A575C7-8917-5101-B8E7-F08ED123A823") {
			// 	ioutil.WriteFile("./debug.json", []byte(jsonObj), 0666)
			// 	fPln("break")
			// }
			/// DEBUG ///

			/// ----------------------------- ///

			// xmlObj, svApplied, err := cvt.JSON2XML(jsonObj, sv)
			// if err != nil {
			// 	status = http.StatusInternalServerError
			// 	RetSB.Reset()
			// 	RetSB.WriteString(err.Error())
			// 	goto RET
			// }
			// logGrp.Do(obj + ":" + svApplied + " applied")

			/// ----------------------------- ///

			// Trace [cvt.JSON2XML]
			results = jaegertracing.TraceFunction(c, cvt.JSON2XML, jsonObj, sv)
			xmlObj := results[0].Interface().(string)
			if !results[2].IsNil() {
				status = http.StatusInternalServerError
				RetSB.Reset()
				RetSB.WriteString(results[2].Interface().(error).Error())
				goto RET
			}
			logGrp.Do(obj + ":" + results[1].Interface().(string) + " applied")

			/// DEBUG ///
			// if sContains(xmlObj, "A5A575C7-8917-5101-B8E7-F08ED123A823") {
			// 	ioutil.WriteFile("./debug.xml", []byte(xmlObj), 0666)
			// 	fPln("break")
			// }
			/// DEBUG ///

			/// ----------------------------- ///

			RetSB.WriteString(xmlObj)
			RetSB.WriteString("\n")
		}

	RET:
		if status != http.StatusOK {
			Ret = RetSB.String()
			warnGrp.Do(Ret + " --> Failed")
		} else {
			if wrapped {
				Ret = sReplaceAll(out4ret, "~~~", "\n"+RetSB.String())
			} else {
				Ret = RetSB.String()
			}
			logGrp.Do("--> Finish JSON2XML")
		}
		return c.String(status, sTrimRight(Ret, "\n")+"\n")
	})
}
