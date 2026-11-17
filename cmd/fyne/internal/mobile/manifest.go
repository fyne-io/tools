// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobile

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
)

type manifestXML struct {
	Activity activityXML `xml:"application>activity"`
}

type activityXML struct {
	Name     string        `xml:"name,attr"`
	MetaData []metaDataXML `xml:"meta-data"`
}

type metaDataXML struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// manifestLibName parses the AndroidManifest.xml and finds the library
// name of the NativeActivity.
// manifestHasForegroundService reports whether the manifest declares a service with an
// android:foregroundServiceType, which an app does in its own AndroidManifest.xml to use a
// foreground service. Mentions of the attribute in comments are ignored.
func manifestHasForegroundService(data []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return false // the end, or XML that the build rejects later on
		}
		el, ok := tok.(xml.StartElement)
		if !ok || el.Name.Local != "service" {
			continue
		}
		for _, attr := range el.Attr {
			if attr.Name.Local == "foregroundServiceType" {
				return true
			}
		}
	}
}

func manifestLibName(data []byte) (string, error) {
	manifest := new(manifestXML)
	if err := xml.Unmarshal(data, manifest); err != nil {
		return "", err
	}
	if manifest.Activity.Name != "org.golang.app.GoNativeActivity" {
		return "", fmt.Errorf("can only build an .apk for GoNativeActivity, not %q", manifest.Activity.Name)
	}
	libName := ""
	for _, md := range manifest.Activity.MetaData {
		if md.Name == "android.app.lib_name" {
			libName = md.Value
			break
		}
	}
	if libName == "" {
		return "", errors.New("AndroidManifest.xml missing meta-data android.app.lib_name")
	}
	return libName, nil
}
