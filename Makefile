VERSION = 1.1.2

APP      := slack-emoji-upload
PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

.PHONY: clean release

clean:
	rm -rf binaries/
	rm -rf release/

release: zip README.md
	git add README.md
	git add Makefile
	git commit -am "Release $(VERSION)" || true
	git push
	hub release create $(VERSION) -m "$(VERSION)" -a release/$(APP)_$(VERSION)_osx_x86_64.zip -a release/$(APP)_$(VERSION)_windows_x86_64.zip -a release/$(APP)_$(VERSION)_linux_x86_64.zip -a release/$(APP)_$(VERSION)_osx_x86_32.zip -a release/$(APP)_$(VERSION)_windows_x86_32.zip -a release/$(APP)_$(VERSION)_linux_x86_32.zip -a release/$(APP)_$(VERSION)_linux_arm64.zip

docker-build-cron: $(GOFILES)
	docker build -t slack-emoji-upload-cron -f contrib/cron/Dockerfile .

zip: release/$(APP)_$(VERSION)_osx_x86_64.zip release/$(APP)_$(VERSION)_windows_x86_64.zip release/$(APP)_$(VERSION)_linux_x86_64.zip release/$(APP)_$(VERSION)_osx_x86_32.zip release/$(APP)_$(VERSION)_windows_x86_32.zip release/$(APP)_$(VERSION)_linux_x86_32.zip release/$(APP)_$(VERSION)_linux_arm64.zip

binaries: binaries/osx_x86_64/$(APP) binaries/windows_x86_64/$(APP).exe binaries/linux_x86_64/$(APP) binaries/osx_x86_32/$(APP) binaries/windows_x86_32/$(APP).exe binaries/linux_x86_32/$(APP)

release/$(APP)_$(VERSION)_osx_x86_64.zip: binaries/osx_x86_64/$(APP)
	mkdir -p release
	cd ./binaries/osx_x86_64 && zip -r -D ../../release/$(APP)_$(VERSION)_osx_x86_64.zip $(APP)

binaries/osx_x86_64/$(APP): $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_64/$(APP)

release/$(APP)_$(VERSION)_windows_x86_64.zip: binaries/windows_x86_64/$(APP).exe
	mkdir -p release
	cd ./binaries/windows_x86_64 && zip -r -D ../../release/$(APP)_$(VERSION)_windows_x86_64.zip $(APP).exe

binaries/windows_x86_64/$(APP).exe: $(GOFILES)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_64/$(APP).exe

release/$(APP)_$(VERSION)_linux_x86_64.zip: binaries/linux_x86_64/$(APP)
	mkdir -p release
	cd ./binaries/linux_x86_64 && zip -r -D ../../release/$(APP)_$(VERSION)_linux_x86_64.zip $(APP)

binaries/linux_x86_64/$(APP): $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_64/$(APP)

release/$(APP)_$(VERSION)_osx_x86_32.zip: binaries/osx_x86_32/$(APP)
	mkdir -p release
	cd ./binaries/osx_x86_32 && zip -r -D ../../release/$(APP)_$(VERSION)_osx_x86_32.zip $(APP)

binaries/osx_x86_32/$(APP): $(GOFILES)
	GOOS=darwin GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/osx_x86_32/$(APP)

release/$(APP)_$(VERSION)_windows_x86_32.zip: binaries/windows_x86_32/$(APP).exe
	mkdir -p release
	cd ./binaries/windows_x86_32 && zip -r -D ../../release/$(APP)_$(VERSION)_windows_x86_32.zip $(APP).exe

binaries/windows_x86_32/$(APP).exe: $(GOFILES)
	GOOS=windows GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/windows_x86_32/$(APP).exe

release/$(APP)_$(VERSION)_linux_x86_32.zip: binaries/linux_x86_32/$(APP)
	mkdir -p release
	cd ./binaries/linux_x86_32 && zip -r -D ../../release/$(APP)_$(VERSION)_linux_x86_32.zip $(APP)

binaries/linux_x86_32/$(APP): $(GOFILES)
	GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_x86_32/$(APP)

release/$(APP)_$(VERSION)_linux_arm64.zip: binaries/linux_arm64/$(APP)
	mkdir -p release
	cd ./binaries/linux_arm64 && zip -r -D ../../release/$(APP)_$(VERSION)_linux_arm64.zip $(APP)

binaries/linux_arm64/$(APP): $(GOFILES)
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o binaries/linux_arm64/$(APP)
