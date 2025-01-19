package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

var BIN_ARCH_NAME_MAP = map[string][]string{
	"linux-amd64": {
		"x86_64-unknown-linux-gnu",
		"x86_64-unknown-linux-gnux32",
		"x86_64-unknown-linux-musl",
		"x86_64-unknown-linux-ohos",
	},
	"linux-arm64": {
		"aarch64-linux-android",
		"aarch64-unknown-linux-gnu",
		"aarch64-unknown-linux-musl",
		"aarch64-unknown-linux-ohos",
	},

	"darwin-arm64": {"aarch64-apple-darwin"},
	"darwin-amd64": {"x86_64-apple-darwin"},

	"windows-arm64": {
		"aarch64-pc-windows-msvc",
		"i586-pc-windows-msvc",
		"i686-pc-windows-gnu",
		"i686-pc-windows-gnullvm",
		"i686-pc-windows-msvc",
		"x86_64-pc-windows-gnu",
		"x86_64-pc-windows-gnullvm",
		"x86_64-pc-windows-msvc",
	},
}

func GetCurrentVersion() string {
	r, err := os.ReadFile("version")
	if err != nil {
		panic("couldn't read file 'version'")
	}

	v := string(r)
	v = strings.ToLower(v)
	v = strings.Trim(v, " ")
	v = strings.Trim(v, fmt.Sprintln())

	return v
}

func GetOutDir() string {
	return path.Join("bin")
}

func GetOutFilePath(goos string, goarch string) string {
	outFile := path.Join("bin", fmt.Sprintf("codeconts-%s-%s", goos, goarch))

	return outFile
}

func Build(goos string, goarch string, version string) {
	fmt.Println("Starting build for", goos, goarch)

	GOOS := strings.ToLower(goos)
	GOARCH := strings.ToLower(goarch)

	name := fmt.Sprintf("%s-%s", GOOS, GOARCH)
	outFile := path.Join("bin", name)

	cmd := exec.Command(
		"go",
		"build",
		"-ldflags",
		fmt.Sprintf("-X 'github.com/fossMeDaddy/codeconts/globals.Version=%s' -s -w", version),
		"-o",
		outFile,
		"main.go",
	)
	cmd.Env = append(os.Environ(), "GOOS="+GOOS, "GOARCH="+GOARCH)
	cmd.Stdout = os.Stdout
	cmdErr := cmd.Run()
	if cmdErr != nil {
		fmt.Println("ERR:", cmdErr)
		panic("error occured while running 'go build'")
	}

	contents_b, readFileErr := os.ReadFile(outFile)
	if readFileErr != nil {
		panic(fmt.Sprintf("error occured while reading file %s", outFile))
	}
	for _, target := range BIN_ARCH_NAME_MAP[name] {
		targetFilePath := path.Join("bin", target)
		if err := os.WriteFile(targetFilePath, contents_b, 0666); err != nil {
			fmt.Println("WARNING: error occured while writing", targetFilePath, err)
		}
	}
	os.Remove(outFile)

	fmt.Println("Built binary for", goos, goarch)
	fmt.Println()
}

func BuildAllTargets(version string) {
	var wg sync.WaitGroup

	wg.Add(3)
	go (func() {
		defer wg.Done()
		Build("linux", "arm64", version)
	})()
	go (func() {
		defer wg.Done()
		Build("linux", "amd64", version)
	})()
	go (func() {
		defer wg.Done()
		Build("linux", "386", version)
	})()

	wg.Add(2)
	go (func() {
		defer wg.Done()
		Build("darwin", "arm64", version)
	})()
	go (func() {
		defer wg.Done()
		Build("darwin", "amd64", version)
	})()

	wg.Add(2)
	go (func() {
		defer wg.Done()
		Build("windows", "amd64", version)
	})()
	go (func() {
		defer wg.Done()
		Build("windows", "386", version)
	})()

	wg.Wait()
}

func CreateTag(version string) {
	// create git tag
	tagCmd := exec.Command("git", "tag", version)
	tagErr := tagCmd.Run()
	if tagErr != nil {
		fmt.Println("ERROR creating new tag:", tagErr)
		return
	}

	// push git tag
	pushCmd := exec.Command("git", "push", "origin", version)
	pushErr := pushCmd.Run()
	if pushErr != nil {
		fmt.Println("ERROR pushing tag:", pushErr)
		return
	}

	// after pushing tag, GH action will build & publish the release for the tag
}
