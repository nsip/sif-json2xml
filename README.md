XXX Copyright, License and other Template necessities

# Web-Service for Converting SIF JSON Format to XML Format

## Build Prerequisite

0. Except 'config.toml' content, Do NOT change any project structure & file name & content.

1. Make sure current working directory of your command line environment is identical to the directory of this README.md file.
   i.e. under "/sif-json2xml/"

## Native Build

0. It is NOT supported to make building on Windows OS. If you are using Windows OS, please choose 'Docker Build'.

1. Make sure `golang` dev package & `git` are available on your machine.

2. Run `./build.sh` to build service which embedded with SIF Spec 3.4.6 & 3.4.7.

3. Run `./release.sh [linux64|win64|mac] 'dest-path'` to extract minimal executable package on different.
   e.g. `./release.sh win64 ~/Desktop/sif-json2xml/` extracts windows version bin package into "~/Desktop/sif-json2xml/".
   then 'server' executable is available under "~/Desktop/sif-json2xml/".

4. Jump into "~/Desktop/sif-json2xml/", modify 'config.toml' if needed.
   Please set [Service] & [Version] to your own value.

5. Run `server`.
   Default port is `1325`, can be set at config.toml.

## Docker Build

0. Make sure `Docker` is available and running on your machine.

1. Run `docker build --rm -t nsip/sif-json2xml:latest .` to make docker image.

2. In order to do configuration before running docker image.
   Copy '/sif-json2xml/config/config.toml' to current directory, modify if needed, and name it like `config_d.toml`.
   Please set [Service] & [Version] to your own value.

3. Run `docker run --rm --mount type=bind,source=$(pwd)/config_d.toml,target=/config.toml -p 0.0.0.0:1325:1325 nsip/sif-json2xml`.
   Default port is `1325`, can be set at config.toml. If not 1325, change above command's '1325' to your own number.

## Test

0. Make sure `curl` is available on your machine.

1. Run `curl IP:Port` to get the list of all available API path of sif-json2xml.
   `IP` : your sif-json2xml server running machine ip.
   `Port`: set in 'config.toml' file, default is 1325, can be changed in 'config.toml'.

2. Run `curl -X POST IP:Port/Service/Version/convert?sv=3.4.7 -d @path/to/your/sif.json`
   to convert a SIF.json to SIF.xml

   `IP` : your sif-json2xml server running machine ip.
   `Port`: Get from server's 'config.toml'-[WebService]-[Port], default is 1325.
   `Service`: service name. Get from server's 'config.toml'-[Service].
   `Version`: service version. Get from server's 'config.toml'-[Version].
   `sv`: SIF Spec Version, available 3.4.6 & 3.4.7
   `wrap`: if there is a single wrapper (non-sif-object root) on upload sif.json, append param `wrap`.  
