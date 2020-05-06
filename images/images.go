package images

import (
	"errors"
	"fmt"
	"github.com/mhewedy/vermin/command"
	"github.com/mhewedy/vermin/db"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func Create(image string) error {
	// check image against cached
	cached, err := listCachedImages()
	if err != nil {
		return err
	}

	for i := range cached {
		if cached[i] == image {
			return nil // already cached
		}
	}

	remote, err := listRemoteImages(false)
	if err != nil {
		return err
	}

	// check image against remote
	var dbImage *dbImage
	for i := range remote {
		r := remote[i]
		if r.Name == image {
			dbImage = &r
			break
		}
	}

	if dbImage == nil {
		l, _ := List()
		return errors.New(fmt.Sprintf("invalid image name: '%s',", image) +
			" valid images are:\n" + strings.Join(l, "\n"))
	}

	return download(dbImage)
}

func CanMount(image string) bool {

	remote, _ := listRemoteImages(false)

	for i := range remote {
		r := remote[i]
		if r.Name == image {
			return r.Mount
		}
	}

	return false
}

type image struct {
	name   string
	cached bool
}

func List() ([]string, error) {
	list, err := list(false)
	if err != nil {
		return nil, err
	}
	var result = make([]string, len(list))
	for i := range list {
		result[i] = list[i].name
	}
	return result, nil
}

func Display(purgeCache bool) (string, error) {

	list, err := list(purgeCache)
	if err != nil {
		return "", err
	}
	var result string

	for i := range list {
		if list[i].cached {
			result += list[i].name + "\t\t(cached)\n"
		} else {
			result += list[i].name + "\n"
		}
	}
	return result, nil
}

func list(purgeCache bool) ([]image, error) {
	var result []image

	cached, err := listCachedImages()
	if err != nil {
		return nil, err
	}
	for i := range cached {
		result = append(result, image{cached[i], true})
	}

	remote, err := listRemoteImagesNames(purgeCache)
	if err != nil {
		return nil, err
	}
	for i := range remote {
		r := remote[i]
		if !contains(cached, r) {
			result = append(result, image{r, false})
		}
	}
	return result, nil
}

func download(r *dbImage) error {
	fmt.Printf("Image '%s' could not be found. Attempting to find and install ...\n", r.Name)

	// download to a temp file
	tmpFile, err := ioutil.TempFile("", strings.ReplaceAll(r.Name, "/", "_"))
	if err != nil {
		return err
	}
	_ = tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	msg := fmt.Sprintf("Downloading: %s", r.URL)
	if _, err := command.Wget(r.URL, tmpFile.Name()).CallWithProgress(msg); err != nil {
		return err
	}

	// copy the downloaded file to images directory
	if err := os.MkdirAll(db.ImagesDir+"/"+strings.Split(r.Name, "/")[0], 0755); err != nil {
		return err
	}
	if err := copyFile(tmpFile.Name(), db.ImagesDir+"/"+r.Name+".ova"); err != nil {
		return err
	}
	return nil
}

func contains(a []string, s string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
