package main

import (
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// 导入
type PsdXML struct {
	XMLName xml.Name `xml:"PSDUI"`
	PsdSize Size     `xml:"psdSize"`
	Layers  Layers   `xml:"layers"`
}

type Size struct {
	Width  int `xml:"width"`
	Height int `xml:"height"`
}
type Layers struct {
	Layer []Layer `xml:"Layer"`
}
type Layer struct {
	Type   string `xml:"type"`
	Name   string `xml:"name"`
	Layers Layers `xml:"layers"`
	Image  Image  `xml:"image"`
}
type Image struct {
	ImageType   string    `xml:"imageType"`
	ImageSource string    `xml:"imageSource"`
	Name        string    `xml:"name"`
	Position    Position  `xml:"position"`
	Size        Size      `xml:"size"`
	Arguments   Arguments `xml:"arguments"`
}
type Position struct {
	X float32 `xml:"x"`
	Y float32 `xml:"y"`
}
type Arguments struct {
	Strings []string `xml:"string"`
}

// 导出
type FUIPackageXML struct {
	XMLName     xml.Name   `xml:"packageDescription"`
	ID          string     `xml:"id,attr"`
	JpegQuality string     `xml:"jpegQuality,attr"`
	CompressPNG string     `xml:"compressPNG,attr"`
	Resources   FResources `xml:"resources"`
}
type FResources struct {
	Images     []FImage    `xml:"image"`
	Components []FResource `xml:"component"`
}
type FResource struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
	Exported string `xml:"exported,attr"`
}
type FImage struct {
	FResource
	Scale      string `xml:"scale,attr"`
	Scale9grid string `xml:"scale9grid,attr"`
}
type FUIComponentXML struct {
	XMLName     xml.Name     `xml:"component"`
	Size        string       `xml:"size,attr"`
	DisplayList FDisplayList `xml:"displayList"`
}
type FDisplayList struct {
	XMLName xml.Name `xml:"displayList"`
	List    []interface{}
}
type FComponentBase struct {
	ID    string `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	XY    string `xml:"xy,attr"`
	Group string `xml:"group,attr"`
}
type FComponent struct {
	XMLName xml.Name `xml:"component"`
	FComponentBase
	Src string `xml:"src,attr"`
}
type FImageComponent struct {
	XMLName xml.Name `xml:"image"`
	FComponentBase
	Src   string `xml:"src,attr"`
	Pivot string `xml:"pivot,attr"`
	Size  string `xml:"size,attr"`
}
type FTextComponent struct {
	XMLName xml.Name `xml:"text"`
	FComponentBase
	Pivot       string `xml:"pivot,attr"`
	Size        string `xml:"size,attr"`
	Input       string `xml:"input,attr"`
	Font        string `xml:"font,attr"`
	FontSize    string `xml:"fontSize,attr"`
	Color       string `xml:"color,attr"`
	StrokeColor string `xml:"strokeColor,attr"`
	StrokeSize  string `xml:"strokeSize,attr"`
	Text        string `xml:"text,attr"`
	Align       string `xml:"align,attr"`
	VAlign      string `xml:"vAlign,attr"`
	AutoSize    string `xml:"autoSize,attr"`
}
type FImageItem struct {
	ID    string
	Image Image
}

func main() {
	inputPath := "D:/psd2ugui/"                                   // PSD导出的文件所在路径，注意最后有个 /
	inputFileName := "psd2ugui"                                   // PSD导出的XML文件名
	outputPath := "E:/testproj/UIProject/UITest/assets/Package1/" // FUI导出目录，注意最后有个 /
	forceFontName := ""                                           // 如果有需要强制使用自己字体的就在这写一下
	file, err := os.Open(inputPath + inputFileName + ".xml")
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	psdXML := PsdXML{}
	err = xml.Unmarshal(data, &psdXML)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	//fmt.Printf("%#v", v)
	totalImageMap := outputPackageXML(psdXML, inputPath, inputFileName, outputPath)
	outputMainComponentXML(psdXML, inputPath, inputFileName, outputPath, totalImageMap, forceFontName)
}
func outputPackageXML(psdXML PsdXML, inputPath string, inputFileName string, outputPath string) map[string]FImageItem {
	// 资源表
	totalImageMap := make(map[string]FImageItem)
	outputXML := FUIPackageXML{}
	outputXML.JpegQuality = "100"
	outputXML.CompressPNG = "true"
	outputXML.ID = "jxk31n82" // 包ID暂时先指定为全相同
	resources := FResources{}
	images := []FImage{}
	components := []FResource{}
	imageIndex := 1
	for _, layer := range psdXML.Layers.Layer {
		for _, inlayer := range layer.Layers.Layer {
			inimage := inlayer.Image
			if inimage.ImageType == "Texture" ||
				inimage.ImageType == "Image" ||
				inimage.ImageType == "SliceImage" {
				_, exist := totalImageMap[inimage.Name]
				if exist {
					continue
				}
				pngName := inimage.Name + ".png"
				// 把文件复制进去
				imageFile, err := os.Open(inputPath + pngName)
				if err != nil {
					fmt.Printf("%s %s\n", inputPath+pngName, err.Error())
					continue
				}
				defer imageFile.Close()
				outputImageFile, err := os.Create(outputPath + pngName)
				if err != nil {
					fmt.Printf("%s %s\n", pngName, err.Error())
					continue
				}
				defer outputImageFile.Close()
				io.Copy(outputImageFile, imageFile)
				imageFile.Seek(0, 0)
				if inimage.ImageType == "Texture" ||
					inimage.ImageType == "Image" {
					nimage := FImage{}
					nimage.ID = "i0" + strconv.Itoa(imageIndex)
					nimage.Name = pngName
					nimage.Path = "/"
					nimage.Scale = "none"
					images = append(images, nimage)
					imageIndex++
					totalImageMap[inimage.Name] = FImageItem{nimage.ID, inimage}
				} else if inimage.ImageType == "SliceImage" {
					nimage := FImage{}
					nimage.ID = "i0" + strconv.Itoa(imageIndex)
					nimage.Name = pngName
					nimage.Path = "/"
					nimage.Scale = "9grid"
					png, _, err := image.Decode(imageFile)
					if err != nil {
						fmt.Printf("%s %s\n", inputPath+nimage.Name, err.Error())
						continue
					}
					left, _ := strconv.Atoi(inimage.Arguments.Strings[0])
					up, _ := strconv.Atoi(inimage.Arguments.Strings[1])
					right, _ := strconv.Atoi(inimage.Arguments.Strings[2])
					down, _ := strconv.Atoi(inimage.Arguments.Strings[3])
					// FUI右边是取宽度减左右的长度
					right = png.Bounds().Max.X - left - right
					//fmt.Printf("%d %d %d\n", png.Bounds().Max.X, left, right)
					down = png.Bounds().Max.Y - up - down
					slice := strconv.Itoa(left) + "," +
						strconv.Itoa(up) + "," +
						strconv.Itoa(right) + "," +
						strconv.Itoa(down)
					nimage.Scale9grid = slice
					images = append(images, nimage)
					imageIndex++
					totalImageMap[inimage.Name] = FImageItem{nimage.ID, inimage}
				}
			}
		}
	}
	// 把组件也加进来
	imageIndex = 1
	ncomp := FResource{}
	ncomp.ID = "t0" + strconv.Itoa(imageIndex)
	ncomp.Name = inputFileName + ".xml" // 主UI名
	ncomp.Path = "/"
	ncomp.Exported = "true"
	components = append(components, ncomp)
	imageIndex++

	resources.Images = images
	resources.Components = components
	outputXML.Resources = resources
	outputText, err := xml.MarshalIndent(outputXML, " ", " ")
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	outputfile, err := os.Create(outputPath + "package.xml")
	if err != nil {
		fmt.Printf("%s", err.Error())
		panic(err)
	}
	defer outputfile.Close()
	outputfile.Write([]byte(xml.Header))
	outputfile.Write(outputText)

	fmt.Println("output package xml done")
	return totalImageMap
}

func outputMainComponentXML(psdXML PsdXML, inputPath string, inputFileName string, outputPath string, totalImageMap map[string]FImageItem, forceFontName string) {
	// 资源表
	outputXML := FUIComponentXML{}
	outputXML.Size = fmt.Sprintf("%d,%d", psdXML.PsdSize.Width, psdXML.PsdSize.Height)
	displayList := FDisplayList{}
	compIndex := 1
	psdWidth := float32(psdXML.PsdSize.Width) / 2.0
	psdHeight := float32(psdXML.PsdSize.Height) / 2.0
	for _, layer := range psdXML.Layers.Layer {
		for _, inlayer := range layer.Layers.Layer {
			inimage := inlayer.Image
			imgWidthDelta := float32(inimage.Size.Width) / 2.0
			imgHeightDelta := float32(inimage.Size.Height) / 2.0
			if inimage.ImageType == "Label" {
				ntext := FTextComponent{}
				ntext.ID = "n" + strconv.Itoa(compIndex) + "_o7fw"
				ntext.Name = inimage.Name
				ntext.Size = fmt.Sprintf("%d,%d", inimage.Size.Width, inimage.Size.Height)
				x, y := inimage.Position.X+psdWidth-imgWidthDelta, psdHeight-inimage.Position.Y-imgHeightDelta
				ntext.Color = "#" + inimage.Arguments.Strings[0]
				if forceFontName != "" {
					ntext.Font = forceFontName
				} else {
					if inimage.Arguments.Strings[1] == "ArialMT" {
						ntext.Font = "Arial"
					} else {
						ntext.Font = inimage.Arguments.Strings[1]
					}
				}
				ntext.FontSize = inimage.Arguments.Strings[2]
				ntext.Text = inimage.Arguments.Strings[3]
				ntext.Align = "center"  // 默认居中
				ntext.VAlign = "middle" // 默认居中
				ntext.AutoSize = "none"
				if len(inimage.Arguments.Strings) > 4 {
					if inimage.Arguments.Strings[4] == "Justification.LEFT" {
						ntext.Align = "left"
					} else if inimage.Arguments.Strings[4] == "Justification.RIGHT" {
						ntext.Align = "right"
						// 居右的时候位置要向右移
						//x += float32(inimage.Size.Width)
					} else {
						// 居中的时候位置要向右移一些
						//x += imgWidthDelta
					}
					ntext.XY = fmt.Sprintf("%.0f,%.0f", x, y)
					if len(inimage.Arguments.Strings) > 5 {
						ntext.StrokeSize = inimage.Arguments.Strings[5]
						ntext.StrokeColor = "#" + inimage.Arguments.Strings[6]
					}
				}
				displayList.List = append(displayList.List, ntext)
				compIndex++
			} else if inimage.ImageType == "Texture" ||
				inimage.ImageType == "Image" ||
				inimage.ImageType == "SliceImage" {
				nimage := FImageComponent{}
				nimage.ID = "n" + strconv.Itoa(compIndex) + "_t45n"
				nimage.Name = inimage.Name
				nimage.Size = fmt.Sprintf("%d,%d", inimage.Size.Width, inimage.Size.Height)
				nimage.Src = totalImageMap[inimage.Name].ID
				nimage.XY = fmt.Sprintf("%.0f,%.0f", inimage.Position.X+psdWidth-imgWidthDelta, psdHeight-inimage.Position.Y-imgHeightDelta)
				//nimage.Group = "none"
				displayList.List = append(displayList.List, nimage)
				compIndex++
			}
		}
	}
	outputXML.DisplayList = displayList

	outputText, err := xml.MarshalIndent(outputXML, " ", " ")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	outputfile, err := os.Create(outputPath + inputFileName + ".xml")
	if err != nil {
		fmt.Printf("%s", err.Error())
		return
	}
	defer outputfile.Close()
	outputfile.Write([]byte(xml.Header))
	outputfile.Write(outputText)

	fmt.Println("output component xml done")
}
