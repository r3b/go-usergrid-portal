GO=go
clean:
	rm -rf bin/
	rm -rf ugportal ugportal~ bindata.go bindata go-bindata ug-portal apigee-portal

goportal: fetch-deps
	#./go-bindata -prefix="portal/2.0.58/bower_components/usergrid-portal/portal/dist/usergrid-portal" portal/2.0.58/bower_components/usergrid-portal/portal/dist/usergrid-portal/...
	./go-bindata -o ug-portal -prefix="usergrid-portal" usergrid-portal/...
	./go-bindata -o apigee-portal -prefix="appsvc-ui" appsvc-ui/...
	go build -o ugportal .
	chmod a+x ugportal
	gzexe ugportal

all: goportal

fetch-deps:
	go get github.com/jteeuwen/go-bindata
	go build -o go-bindata github.com/jteeuwen/go-bindata/go-bindata
	# chmod a+x bindata

.PHONEY: clean fetch-deps all
