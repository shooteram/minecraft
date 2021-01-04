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
	"path/filepath"
	"regexp"
	"runtime"
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

// Classifiers _
type Classifiers struct {
	JavaDoc        Artifact `json:"javadoc"`
	Sources        Artifact `json:"sources"`
	NativesLinux   Artifact `json:"natives-linux"`
	NativesWindows Artifact `json:"natives-windows"`
	NativesMacOs   Artifact `json:"natives-macos"`
	NativesOSX     Artifact `json:"natives-osx"`
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
			Artifact    Artifact    `json:"artifact"`
			Classifiers Classifiers `json:"classifiers"`
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

var rootPath string

func main() {
	var version string
	var server bool
	flag.StringVar(&version, "version", "release", "Codename for the wanted version (e.g., release, snapshot, 1.16.4, 20w51a)")
	flag.BoolVar(&server, "server", false, "Launch server instance. Default launch a client version")
	flag.Parse()

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Could not get the user config dir\n%s", err.Error())
	}
	rootPath := path.Join(userConfigDir, "minecraft")

	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		os.Mkdir(rootPath, 0755)
	}

	if _, err := os.Stat(path.Join(rootPath, "versions")); os.IsNotExist(err) {
		os.Mkdir(path.Join(rootPath, "versions"), 0755)
	}

	var native string
	native = runtime.GOOS
	if runtime.GOOS == "darwin" {
		native = "macos"
	}
	native = fmt.Sprintf("natives-%s", native)
	fmt.Printf("Runtime: %s\n", native)

	_, manifestExists := os.Stat(path.Join(rootPath, "versions", "version_manifest.json"))
	releaseOrSnapshot, err := regexp.Match(`release|snapshot`, []byte(version))
	if err != nil {
		log.Fatalf("Could not get the version manifest\n%s", err.Error())
	}

	var manifest Manifest
	if releaseOrSnapshot || os.IsNotExist(manifestExists) {
		log.Println("fetching version_manifest.json")
		resp, err := http.Get("https://launchermeta.mojang.com/mc/game/version_manifest.json")
		if err != nil {
			log.Fatalf("Could not get the version manifest\n%s", err.Error())
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Could not read the version manifest\n%s", err.Error())
		}

		err = ioutil.WriteFile(path.Join(rootPath, "versions", "version_manifest.json"), body, 0644)
		if err != nil {
			log.Fatalf("Could not write manifest\n%s", err.Error())
		}

		err = json.Unmarshal(body, &manifest)
		if err != nil {
			log.Fatalf("Could not parse the version manifest\n%s", err.Error())
		}
	} else {
		data, err := ioutil.ReadFile(path.Join(rootPath, "versions", "version_manifest.json"))
		if err != nil {
			log.Fatalf("Could not read manifest\n%s", err.Error())
		}

		err = json.Unmarshal(data, &manifest)
		if err != nil {
			log.Fatalf("Could not parse the manifest\n%s", err.Error())
		}
	}

	if "release" == version {
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

	if _, err := os.Stat(path.Join(rootPath, "versions", version)); os.IsNotExist(err) {
		os.Mkdir(path.Join(rootPath, "versions", version), 0755)
	}

	var pkg Package
	if _, err := os.Stat(path.Join(rootPath, "versions", version, "meta.json")); os.IsNotExist(err) {
		log.Printf("fetching %s\n", versionManifest)
		resp, err := http.Get(versionManifest)
		if err != nil {
			log.Fatalf("Could not get the package manifest\n%s", err.Error())
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Could not read the package\n%s", err.Error())
		}

		err = ioutil.WriteFile(path.Join(rootPath, "versions", version, "meta.json"), body, 0644)
		if err != nil {
			log.Fatalf("Could not write package manifest\n%s", err.Error())
		}

		err = json.Unmarshal(body, &pkg)
		if err != nil {
			log.Fatalf("Could not parse the package\n%s", err.Error())
		}
	} else {
		data, err := ioutil.ReadFile(path.Join(rootPath, "versions", version, "meta.json"))
		if err != nil {
			log.Fatalf("Could not read package manifest\n%s", err.Error())
		}

		err = json.Unmarshal(data, &pkg)
		if err != nil {
			log.Fatalf("Could not parse the package\n%s", err.Error())
		}
	}

	libPath := path.Join(rootPath, "versions", version, "libraries")
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		os.Mkdir(libPath, 0755)
	}

	nativeLibPath := path.Join(rootPath, "versions", version, "native-libraries")
	if _, err := os.Stat(nativeLibPath); os.IsNotExist(err) {
		os.Mkdir(nativeLibPath, 0755)
	}

	clientPath := path.Join(rootPath, "versions", version, fmt.Sprintf("client_%s.jar", version))
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		log.Printf("Downloading %s\n", pkg.Downloads.Client.URL)

		resp, err := http.Get(pkg.Downloads.Client.URL)
		if err != nil {
			log.Fatalf("Could not get the client.jar for the version %s\n%s\n", version, err.Error())
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Could not read the bytes of the downloaded client.jar for the version %s\n%s\n", version, err.Error())
		}

		err = ioutil.WriteFile(clientPath, body, 0644)
		if err != nil {
			log.Fatalf("Could not write the client.jar for the version %s\n%s\n", version, err.Error())
		}
	}

	for _, l := range pkg.Libraries {
		if native == l.Natives.Linux || native == l.Natives.Windows || native == l.Natives.OSX {
			var nativeURL string
			var nativeLibraryName string

			switch native {
			case l.Natives.Linux:
				nativeURL = l.Downloads.Classifiers.NativesLinux.URL
				nativeLibraryName = filepath.Base(l.Downloads.Classifiers.NativesLinux.Path)
			case l.Natives.Windows:
				nativeURL = l.Downloads.Classifiers.NativesWindows.URL
				nativeLibraryName = filepath.Base(l.Downloads.Classifiers.NativesWindows.Path)
			case l.Natives.OSX:
				nativeURL = l.Downloads.Classifiers.NativesOSX.URL
				nativeLibraryName = filepath.Base(l.Downloads.Classifiers.NativesOSX.Path)
			}

			nativeLib := path.Join(nativeLibPath, nativeLibraryName)
			if _, err := os.Stat(nativeLib); os.IsNotExist(err) {
				resp, err := http.Get(nativeURL)
				if err != nil {
					log.Fatalf("Could not get the native library %s\n%s", filepath.Base(nativeLib), err.Error())
				}

				body, _ := ioutil.ReadAll(resp.Body)
				_ = ioutil.WriteFile(nativeLib, body, 0644)
			}

			continue
		}

		libdir := path.Join(libPath, filepath.Dir(l.Downloads.Artifact.Path))
		if _, err := os.Stat(libdir); os.IsNotExist(err) {
			err := os.MkdirAll(libdir, 0755)
			if err != nil {
				log.Fatalf("Could not create directory %s\n%s", libdir, err.Error())
			}
		}

		lfile := path.Join(libPath, l.Downloads.Artifact.Path)
		if _, err := os.Stat(lfile); os.IsNotExist(err) {
			resp, err := http.Get(l.Downloads.Artifact.URL)
			if err != nil {
				log.Fatalf("Could not get the artifact %s\n%s", filepath.Base(l.Downloads.Artifact.Path), err.Error())
			}

			body, _ := ioutil.ReadAll(resp.Body)
			_ = ioutil.WriteFile(lfile, body, 0644)
		}
	}
}
