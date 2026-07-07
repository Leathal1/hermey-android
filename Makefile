# Hermdroid Makefile
# Orchestrates: Go core → gomobile AAR → Gradle APK

GO := go
GOMOBILE := gomobile
GRADLE := ./gradlew

.PHONY: all aar apk test clean

all: test aar apk

# ── Go core ──

test:
	cd core && $(GO) test ./... -v -count=1

vet:
	cd core && $(GO) vet ./...

# ── gomobile AAR ──

aar: test
	cd core && $(GOMOBILE) bind -target=android/arm64 -o ../app/libs/hermey-core.aar .

aar-x86: test
	cd core && $(GOMOBILE) bind -target=android/386 -o ../app/libs/hermey-core-x86.aar .

# ── Android APK ──

apk: aar
	cd app && $(GRADLE) assembleDebug

apk-release: aar
	cd app && $(GRADLE) assembleRelease

# ── Clean ──

clean:
	rm -f app/libs/hermey-core.aar app/libs/hermey-core-x86.aar
	cd app && $(GRADLE) clean
