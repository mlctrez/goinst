package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Version is the semver plus release candidate representation of the go version
type Version struct {
	Major            int
	Minor            int
	Patch            int
	ReleaseCandidate bool
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	goos, goarch, err := getOsArch()
	exitOnError(err)

	fmt.Printf("goos=%q goarch=%q\n", goos, goarch)

	versions := getVersions()
	latestVersion := versions[len(versions)-1]

	ul := "/usr/local"

	installPath := path.Join(ul, "go")

	installedVersion, err := getCurrentGoVersion(installPath)
	exitOnError(err)

	if installedVersion.LessThan(latestVersion) {
		fmt.Println("go version", latestVersion.VersionTag(), "is the latest and will be installed")

		// fail fast if writing to /usr/local will fail
		extractDir, err := ioutil.TempDir(ul, "go_install_")
		fmt.Println(extractDir)

		exitOnError(err)
		defer os.RemoveAll(extractDir)

		dv, err := ioutil.TempFile(extractDir, latestVersion.VersionTag()+".tar.gz_")
		exitOnError(err)

		link := latestVersion.DownLoadLink(goos, goarch)
		fmt.Println("downloading", link)
		resp, err := http.Get(link)
		exitOnError(err)

		_, err = io.Copy(dv, resp.Body)
		exitOnError(err)
		dv.Close()
		resp.Body.Close()
		fmt.Println("download complete")

		_, err = exec.Command("tar", "-xf", dv.Name(), "-C", extractDir).CombinedOutput()
		exitOnError(err)

		if pathExists(installPath) {
			exitOnError(os.Rename(installPath, path.Join(ul, installedVersion.VersionTag())))
		}
		exitOnError(os.Rename(path.Join(extractDir, "go"), installPath))

	} else {
		fmt.Println("go version", installedVersion.VersionTag(), "is the latest and is currently installed")
	}

	fmt.Println("echo -e '# Expand the $PATH to include /usr/local/go/bin \\nPATH=$PATH:/usr/local/go/bin'  >> /etc/profile.d/golang.sh")

}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return !os.IsNotExist(err)
}

func getCurrentGoVersion(installPath string) (v *Version, err error) {
	if pathExists(installPath) {
		output, err := exec.Command(path.Join(installPath, "bin/go"), "version").CombinedOutput()
		if err != nil {
			panic(err)
		}
		vp := strings.Split(string(output), " ")
		version := strings.TrimPrefix(vp[2], "go")
		return ParseVersion(version)
	}
	return ParseVersion("0.0.0")
}

func getOsArch() (os, arch string, err error) {
	var output []byte
	if output, err = exec.Command("uname", "-s").CombinedOutput(); err != nil {
		return
	}
	os = strings.ToLower(strings.TrimSpace(string(output)))
	if output, err = exec.Command("uname", "-m").CombinedOutput(); err != nil {
		return
	}
	arch = strings.ToLower(strings.TrimSpace(string(output)))
	if arch == "x86_64" {
		arch = "amd64"
	}
	if arch == "i686" {
		arch = "386"
	}
	return
}

func getVersions() []*Version {

	res, err := http.Get("https://golang.org/dl/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	versions := make([]*Version, 0)

	doc.Find(".download").Each(func(i int, s *goquery.Selection) {
		if h, ok := s.Attr("href"); ok {
			// https://dl.google.com/go/go1.3.src.tar.gz -> 1.3
			if strings.HasSuffix(h, ".src.tar.gz") {
				gv := strings.TrimPrefix(h, "/dl/go")
				gv = strings.TrimSuffix(gv, ".src.tar.gz")
				v, e := ParseVersion(gv)
				if e == nil {
					versions = append(versions, v)
				}
			}
		}
	})

	sort.Sort(ByVersion(versions))
	return versions
}

// VersionTag produces a go version tag, dropping trailing zero parts
func (v *Version) VersionTag() string {
	u := fmt.Sprintf("go%d", v.Major)

	if v.Minor > 0 || v.Patch > 0 {
		u += fmt.Sprintf(".%d", v.Minor)
	}
	if v.Patch > 0 {
		patchFormat := ".%d"
		if v.ReleaseCandidate {
			patchFormat = "rc%d"
		}
		u += fmt.Sprintf(patchFormat, v.Patch)
	}
	return u
}

// DownLoadLink calculates the download link based on os and arch
func (v *Version) DownLoadLink(os, arch string) string {
	u := "https://dl.google.com/go/"

	u += v.VersionTag()
	u += "." + os + "-" + arch
	if os == "linux" || os == "freebsd" || os == "darwin" {
		u += ".tar.gz"
	}
	if os == "windows" {
		u += ".zip"
	}
	return u
}

/*

  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.src.tar.gz">go1.10.2.src.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.darwin-amd64.tar.gz">go1.10.2.darwin-amd64.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.darwin-amd64.pkg">go1.10.2.darwin-amd64.pkg</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-386.tar.gz">go1.10.2.linux-386.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz">go1.10.2.linux-amd64.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-armv6l.tar.gz">go1.10.2.linux-armv6l.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.windows-386.zip">go1.10.2.windows-386.zip</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.windows-386.msi">go1.10.2.windows-386.msi</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.windows-amd64.zip">go1.10.2.windows-amd64.zip</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.windows-amd64.msi">go1.10.2.windows-amd64.msi</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.freebsd-386.tar.gz">go1.10.2.freebsd-386.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.freebsd-amd64.tar.gz">go1.10.2.freebsd-amd64.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-arm64.tar.gz">go1.10.2.linux-arm64.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-ppc64le.tar.gz">go1.10.2.linux-ppc64le.tar.gz</a></td>
  <td class="filename"><a class="download" href="https://dl.google.com/go/go1.10.2.linux-s390x.tar.gz">go1.10.2.linux-s390x.tar.gz</a></td>

*/

// LessThan for comparing versions
func (v *Version) LessThan(o *Version) bool {
	if v.Major != o.Major {
		return v.Major < o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor < o.Minor
	}

	if v.ReleaseCandidate && o.ReleaseCandidate {
		return v.Patch < o.Patch
	}
	if !v.ReleaseCandidate && !o.ReleaseCandidate {
		return v.Patch < o.Patch
	}
	if v.ReleaseCandidate && !o.ReleaseCandidate {
		return true
	}
	if !v.ReleaseCandidate && o.ReleaseCandidate {
		return false
	}

	return v.Patch < o.Patch
}

// ByVersion allows for sorting
type ByVersion []*Version

// Len for sort
func (a ByVersion) Len() int { return len(a) }

// Swap for sort
func (a ByVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less for sort
func (a ByVersion) Less(i, j int) bool { return a[i].LessThan(a[j]) }

// ParseVersion string -> int,int,int semver
func ParseVersion(ver string) (v *Version, err error) {

	v = &Version{}

	if strings.Contains(ver, "rc") {
		v.ReleaseCandidate = true
		ver = strings.Replace(ver, "rc", ".", 1)
	}

	p := strings.Split(ver, ".")

	if len(p) > 0 {
		i, e := strconv.Atoi(p[0])
		if e != nil {
			return v, e
		}
		v.Major = i
	}
	if len(p) > 1 {
		i, e := strconv.Atoi(p[1])
		if e != nil {
			return v, e
		}
		v.Minor = i
	}
	if len(p) > 2 {
		i, e := strconv.Atoi(p[2])
		if e != nil {
			return v, e
		}
		v.Patch = i
	}

	return
}
