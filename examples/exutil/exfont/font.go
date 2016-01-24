package exfont

import (
	"io/ioutil"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/mobile/asset"
)

// LoadAsset loads the asset at path and interprets it as a font for rendering
// with golang.org/x/image/font using opt to create the font.Face object.
func LoadAsset(path string, opt *truetype.Options) (*truetype.Font, font.Face, error) {
	f, err := asset.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}
	ttf, err := freetype.ParseFont(raw)
	if err != nil {
		return nil, nil, err
	}
	face := truetype.NewFace(ttf, opt)
	return ttf, face, nil
}
