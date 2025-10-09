package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
)

type Color struct {
    Red, Green, Blue, Alpha int
}

type OctreeNode struct {
    Color        Color
    PixelCount   int
    PaletteIndex int
    Children     [8]*OctreeNode
}

const MaxDepth = 8

type OctreeQuantizer struct {
	// this is probably to get what level the octree node is on.
    Levels map[int][]*OctreeNode
    Root   *OctreeNode
}

func NewColor(red, green, blue, alpha int) Color {
    return Color{Red: red, Green: green, Blue: blue, Alpha: alpha}
}

func NewOctreeNode(level int, parent *OctreeQuantizer) *OctreeNode {
    node := &OctreeNode{
        Color: Color{0, 0, 0, 0},
    }
    if level < MaxDepth-1 {
        parent.AddLevelNode(level, node)
    }
    return node
}

func (node *OctreeNode) IsLeaf() bool {
    return node.PixelCount > 0
}

func (node *OctreeNode) GetLeafNodes() []*OctreeNode {
    var leafNodes []*OctreeNode
    for _, child := range node.Children {
        if child != nil {
            if child.IsLeaf() {
                leafNodes = append(leafNodes, child)
            } else {
                leafNodes = append(leafNodes, child.GetLeafNodes()...)
            }
        }
    }
    return leafNodes
}

func (node *OctreeNode) GetNodesPixelCount() int {
    sumCount := node.PixelCount
    for _, child := range node.Children {
        if child != nil {
            sumCount += child.PixelCount
        }
    }
    return sumCount
}

func (node *OctreeNode) AddColor(color Color, level int, parent *OctreeQuantizer) {
    if level >= MaxDepth {
        node.Color.Red += color.Red
        node.Color.Green += color.Green
        node.Color.Blue += color.Blue
        node.Color.Alpha += color.Alpha
        node.PixelCount++
        return
    }
    index := node.GetColorIndexForLevel(color, level)
	// the deref happens here!
    if node.Children[index] == nil {
        node.Children[index] = NewOctreeNode(level, parent)
    }
    node.Children[index].AddColor(color, level+1, parent)
}

func (node *OctreeNode) GetPaletteIndex(color Color, level int) int {
    if node.IsLeaf() {
        return node.PaletteIndex
    }
    index := node.GetColorIndexForLevel(color, level)
    if node.Children[index] != nil {
        return node.Children[index].GetPaletteIndex(color, level+1)
    }
    for _, child := range node.Children {
        if child != nil {
            return child.GetPaletteIndex(color, level+1)
        }
    }
    return 0
}

func (node *OctreeNode) RemoveLeaves() int {
    result := 0
    for i := range node.Children {
        if node.Children[i] != nil {
            node.Color.Red += node.Children[i].Color.Red
            node.Color.Green += node.Children[i].Color.Green
            node.Color.Blue += node.Children[i].Color.Blue
            node.Color.Alpha += node.Children[i].Color.Alpha
            node.PixelCount += node.Children[i].PixelCount
            result++
        }
		node.Children[i] = nil;
    }
    return result - 1
}

func (node *OctreeNode) GetColorIndexForLevel(color Color, level int) int {
    index := 0
    mask := 0x80 >> level
    if color.Red&mask != 0 {
        index |= 4
    }
    if color.Green&mask != 0 {
        index |= 2
    }
    if color.Blue&mask != 0 {
        index |= 1
    }
    return index
}

func (node *OctreeNode) GetColor() Color {
    if node.PixelCount == 0 {
        return Color{0, 0, 0, 0}
    }
    return Color{
        Red:   node.Color.Red / node.PixelCount,
        Green: node.Color.Green / node.PixelCount,
        Blue:  node.Color.Blue / node.PixelCount,
        Alpha: node.Color.Alpha / node.PixelCount,
    }
}

func NewOctreeQuantizer() *OctreeQuantizer {
    quantizer := &OctreeQuantizer{
        Levels: make(map[int][]*OctreeNode),
    }
    quantizer.Root = NewOctreeNode(0, quantizer)
    return quantizer
}

func (quantizer *OctreeQuantizer) GetLeaves() []*OctreeNode {
    return quantizer.Root.GetLeafNodes()
}

func (quantizer *OctreeQuantizer) AddLevelNode(level int, node *OctreeNode) {
    quantizer.Levels[level] = append(quantizer.Levels[level], node)
}

func (quantizer *OctreeQuantizer) AddColor(color Color) {
    quantizer.Root.AddColor(color, 0, quantizer)
}

func (quantizer *OctreeQuantizer) MakePalette(colorCount int) []Color {
    var palette []Color
    paletteIndex := 0
    leafCount := len(quantizer.GetLeaves())
	fmt.Printf("Before removal, Length of leaves: %d\n", leafCount);
    for level := MaxDepth - 1; level >= 0; level-- {
        if nodes, exists := quantizer.Levels[level]; exists {
            for _, node := range nodes {
                leafCount -= node.RemoveLeaves()
                if leafCount <= colorCount {
                    break
                }
            }
            if leafCount <= colorCount {
                break
            }
            quantizer.Levels[level] = nil
        }
    }
	leaves := quantizer.GetLeaves();
	fmt.Printf("After removal, Length of leaves: %d\n", len(leaves));
    for _, node := range leaves {
        if paletteIndex >= colorCount {
            break
        }
        if node.IsLeaf() {
            palette = append(palette, node.GetColor())
            node.PaletteIndex = paletteIndex
            paletteIndex++
        }
    }
    return palette
}

func (quantizer *OctreeQuantizer) GetPaletteIndex(color Color) int {
    return quantizer.Root.GetPaletteIndex(color, 0)
}

// very slow on tight loops, do not use this in animated gifs.
func (quantizer *OctreeQuantizer) ConvertRGBAToPalettedImage(img *image.RGBA, palette color.Palette) image.PalettedImage  {
	paletted_image := image.NewPaletted(img.Rect, palette);
	bounds := img.Bounds();

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := y*img.Stride + x*4
			r := img.Pix[i+0]
			g := img.Pix[i+1]
			b := img.Pix[i+2]
			a := img.Pix[i+3]
			if a == 0 {
				continue;
			}

			color := NewColor(int(r), int(g), int(b), int(a));
			index := quantizer.GetPaletteIndex(color);
			paletted_image.SetColorIndex(x, y, uint8(index));
		}
	}

	return paletted_image;
}

func ConvertToColorPalette(palette []Color) color.Palette {
    var colorPalette color.Palette
    for _, c := range palette {
        colorPalette = append(colorPalette, color.RGBA{
            R: uint8(c.Red),
            G: uint8(c.Green),
            B: uint8(c.Blue),
            A: uint8(c.Alpha),
        })
    }
    return colorPalette
}

func AddColorsToQuantizer(q *OctreeQuantizer, g *gif.GIF) {
    // Add colors from each frame to the quantizer
    for _, frame := range g.Image {
        bounds := frame.Bounds()
        for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
            for x := bounds.Min.X; x < bounds.Max.X; x++ {
                r, g, b, a := frame.At(x, y).RGBA()
                color := NewColor(int(r>>8), int(g>>8), int(b>>8), int(a>>8))
                q.AddColor(color) // called every pixel in every frame!
            }
        }
    }
}


