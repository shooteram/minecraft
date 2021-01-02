package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
)

// Manifest _
type Manifest struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		ReleaseTime string `json:"releaseTime"`
		Time        string `json:"time"`
		Type        string `json:"type"`
		URL         string `json:"url"`
	} `json:"versions"`
}

// Download _
type Download struct {
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// Artifact _
type Artifact struct {
	Path string `json:"path"`
	SHA1 string `json:"sha1"`
	Size int    `json:"size"`
	URL  string `json:"url"`
}

// Natives _
type Natives struct {
	Linux   string `json:"linux"`   // natives-linux
	Windows string `json:"windows"` // natives-windows
	OSX     string `json:"osx"`     // natives-macos
}

// Package _
type Package struct {
	Arguments struct {
		Game json.RawMessage `json:"game"`
		JVM  json.RawMessage `json:"jvm"`
	} `json:"arguments"`
	AssetIndex struct {
		ID        string `json:"id"`
		SHA1      string `json:"sha1"`
		Size      int    `json:"size"`
		TotalSize int    `json:"totalSize"`
		URL       string `json:"url"`
	} `json:"assetIndex"`
	Assets          string `json:"assets"`
	ComplianceLevel int    `json:"complianceLevel"`
	Downloads       struct {
		Client Download `json:"client"`
		Server Download `json:"server"`
	} `json:"downloads"`
	ID        string `json:"id"`
	Libraries []struct {
		Downloads struct {
			Artifact    Artifact `json:"artifact"`
			Classifiers struct {
				JavaDoc        Artifact `json:"javadoc"`
				Sources        Artifact `json:"sources"`
				NativesLinux   Artifact `json:"natives-linux"`
				NativesWindows Artifact `json:"natives-windows"`
				NativesMacOs   Artifact `json:"natives-macos"`
				NativesOSX     Artifact `json:"natives-osx"`
			} `json:"classifiers"`
		} `json:"downloads"`
		Extract struct {
			Exclude []string `json:"exclude"`
		} `json:"extract"`
		Name    string  `json:"name"`
		Natives Natives `json:"natives"`
		Rules   []struct {
			Action string `json:"action"`
			OS     struct {
				Name string `json:"name"`
			} `json:"os"`
		} `json:"rules"`
	} `json:"libraries"`
}

// X _
type X struct {
}

func main() {
	var version string
	var server bool
	flag.StringVar(&version, "version", "", "Codename for the wanted version (e.g., release, snapshot, 1.16.4, 20w51a)")
	flag.BoolVar(&server, "server", false, "Launch server instance. Default launch a client version")
	flag.Parse()

	if _, err := os.Stat("versions"); os.IsNotExist(err) {
		os.Mkdir("versions", 0755)
	}

	var nativeOS string
	nativeOS = runtime.GOOS
	if runtime.GOOS == "darwin" {
		nativeOS = "macos"
	}
	fmt.Printf("Runtime: %s\n", nativeOS)

	if _, err := os.Stat(path.Join("versions", "version_manifest.json")); os.IsNotExist(err) {
	}

	resp, err := http.Get("https://launchermeta.mojang.com/mc/game/version_manifest.json")
	if err != nil {
		log.Fatalf("Could not get the version manifest\n%s\n", err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Could not read the version manifest\n%s\n", err.Error())
	}

	var manifest Manifest
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		log.Fatalf("Could not parse the version manifest\n%s\n", err.Error())
	}

	if version == "" {
		version = manifest.Latest.Release
	}

	if "snapshot" == version {
		version = manifest.Latest.Snapshot
	}

	var versionManifest string
	for _, v := range manifest.Versions {
		if version == v.ID {
			versionManifest = v.URL
			break
		}
	}

	if versionManifest == "" {
		log.Fatalf("Did not find the manifest url for the requested version: %s\n", version)
	}

	if _, err := os.Stat(path.Join("versions", version)); os.IsNotExist(err) {
		os.Mkdir(path.Join("versions", version), 0755)
	}

	var pkg Package
	if _, err := os.Stat(path.Join("versions", version, "meta.json")); os.IsNotExist(err) {
		resp, err = http.Get(versionManifest)
		if err != nil {
			log.Fatalf("Could not- get the package manifest\n%s\n", err.Error())
		}

		if err != nil {
			log.Fatalf("Could not read the package\n%s\n", err.Error())
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Could not read the package\n%s\n", err.Error())
		}

		err = ioutil.WriteFile(path.Join("versions", version, "meta.json"), body, 0644)
		if err != nil {
			log.Fatalf("Could not write package manifest\n%s\n", err.Error())
		}

		err = json.Unmarshal(body, &pkg)
		if err != nil {
			log.Fatalf("Could not parse the package\n%s\n", err.Error())
		}
	} else {
		data, err := ioutil.ReadFile(path.Join("versions", version, "meta.json"))
		if err != nil {
			log.Fatalf("Could not read package manifest\n%s\n", err.Error())
		}

		err = json.Unmarshal(data, &pkg)
		if err != nil {
			log.Fatalf("Could not parse the package\n%s\n", err.Error())
		}
	}

	fmt.Printf("%s: %s\n", pkg.ID, pkg.Downloads.Client.URL)
	fmt.Printf("\nLibraries\n")

	for _, l := range pkg.Libraries {
		fmt.Printf("\t- %s (v%s)\n", strings.Split(l.Name, ":")[1], strings.Split(l.Name, ":")[2])
	}
}
