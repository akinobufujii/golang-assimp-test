package main

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/andrebq/assimp/conv"
	"github.com/go-gl/gl/all-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// 頂点シェーダプログラム
var vertexShader = `
#version 410

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 in_position;
in vec4 in_vertexColor;
out vec4 out_vertexColor;

void main() {
	gl_Position = projection * camera * model * vec4(in_position, 1);
	out_vertexColor = in_vertexColor;
}
` + "\x00"

// フラグメントシェーダプログラム
var fragmentShader = `
#version 410

in vec4 out_vertexColor;
out vec4 out_color;

void main() {
	out_color = out_vertexColor;
}
` + "\x00"

// MeshVertexFormat メッシュ用頂点フォーマット
type MeshVertexFormat struct {
	mPos [3]float32 // 位置
	//	mUV     assimp.Vector2 // UV座標
	//	mNormal assimp.Vector3 // 法線
	mColor [4]float32 // 色
}

// MeshData メッシュデータ構造体
type MeshData struct {
	mVertexData  []MeshVertexFormat // 頂点データ
	mIndices     []uint32           // インデックスデータ
	mTextureName string             // テクスチャ名
}

// Model モデル構造体
type Model struct {
	mMeshDataList []MeshData // メッシュデータリスト
}

// LoadModel モデルデータ読み込み
func (model *Model) LoadModel(path string) {
	// シーン読み込み
	scene, err := conv.LoadAsset(path)
	if err != nil {
		fmt.Printf("Unable to load scene.\nCause: %v", err)
		panic(err)
	}

	model.mMeshDataList = make([]MeshData, len(scene.Mesh))
	for i, meshData := range scene.Mesh {

		colorList := [][]float32{
			{0, 0, 0, 1},
			{1, 0, 0, 1},
			{0, 1, 0, 1},
			{0, 0, 1, 1},
			{1, 1, 0, 1},
			{0, 1, 1, 1},
			{1, 0, 1, 1},
			{1, 1, 1, 1}}

		colorIndex := 0
		colorListMax := len(colorList)

		// 頂点情報を格納
		fmt.Println("vertex")
		model.mMeshDataList[i].mVertexData = make([]MeshVertexFormat, len(meshData.Vertices))
		for j := range meshData.Vertices {
			model.mMeshDataList[i].mVertexData[j].mPos[0] = float32(meshData.Vertices[j][0])
			model.mMeshDataList[i].mVertexData[j].mPos[1] = float32(meshData.Vertices[j][1])
			model.mMeshDataList[i].mVertexData[j].mPos[2] = float32(meshData.Vertices[j][2])
			//model.mMeshDataList[i].mVertexData[j].mUV = meshData.UVCoords[j]
			//model.mMeshDataList[i].mVertexData[j].mColor = meshData.Colors[j]
			fmt.Printf("%v %v %v\n",
				model.mMeshDataList[i].mVertexData[j].mPos[0],
				model.mMeshDataList[i].mVertexData[j].mPos[1],
				model.mMeshDataList[i].mVertexData[j].mPos[2])

			// 適当な色を突っ込んでおく
			model.mMeshDataList[i].mVertexData[j].mColor[0] = colorList[colorIndex][0]
			model.mMeshDataList[i].mVertexData[j].mColor[1] = colorList[colorIndex][1]
			model.mMeshDataList[i].mVertexData[j].mColor[2] = colorList[colorIndex][2]
			model.mMeshDataList[i].mVertexData[j].mColor[3] = colorList[colorIndex][3]

			colorIndex++
			if colorIndex >= colorListMax {
				colorIndex = 0
			}
		}

		fmt.Println("array")
		for _, data := range model.mMeshDataList[i].mVertexData {
			fmt.Printf("%v %v %v\n",
				data.mPos[0],
				data.mPos[1],
				data.mPos[2])
		}

		// 添字情報を格納
		fmt.Println("index")
		count := 0
		model.mMeshDataList[i].mIndices = make([]uint32, len(meshData.Faces)*3)
		for _, face := range meshData.Faces {
			for _, index := range face.Indices {
				model.mMeshDataList[i].mIndices[count] = uint32(index)
				fmt.Printf("%v ", model.mMeshDataList[i].mIndices[count])
				count++
			}
			fmt.Println("")
		}

		fmt.Println("array")
		count = 0
		for _, index := range model.mMeshDataList[i].mIndices {
			fmt.Printf("%v ", index)

			count++
			if count >= 3 {
				count = 0
				fmt.Println("")
			}
		}
	}
}

// 初期化関数
func init() {
	// 必ずメインスレッドで呼ぶ必要がある
	runtime.LockOSThread()
}

func main() {

	// GLの初期化
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate() // 終了時呼び出し

	// GLFWの設定
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Window作成
	window, err := glfw.CreateWindow(1280, 960, "3D Model Viewer", nil, nil)
	if err != nil {
		panic(err)
	}

	// カレントコンテキスト作成
	window.MakeContextCurrent()

	// GLの初期化
	if err := gl.Init(); err != nil {
		panic(err)
	}
	fmt.Println("OpenGL version", gl.GoStr(gl.GetString(gl.VERSION)))

	// シェーダプログラム作成
	program, err := CreateShaderProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)
	gl.BindFragDataLocation(program, 0, gl.Str("out_color\x00"))

	// モデルデータ読み込み
	sampleModel := new(Model)
	sampleModel.LoadModel("models/gun.obj")

	// 頂点情報作成
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	defer gl.BindVertexArray(0)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	dataSize := len(sampleModel.mMeshDataList[0].mVertexData) * int(unsafe.Sizeof(sampleModel.mMeshDataList[0].mVertexData[0]))
	fmt.Printf("dataSize = %v\n", dataSize)
	gl.BufferData(gl.ARRAY_BUFFER, dataSize, gl.Ptr(sampleModel.mMeshDataList[0].mVertexData), gl.STATIC_DRAW)

	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	dataSize = len(sampleModel.mMeshDataList[0].mIndices) * int(unsafe.Sizeof(sampleModel.mMeshDataList[0].mIndices[0]))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, dataSize, gl.Ptr(sampleModel.mMeshDataList[0].mIndices), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("in_position\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 4*7, gl.PtrOffset(0))

	vertAttrib = uint32(gl.GetAttribLocation(program, gl.Str("in_vertexColor\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 4, gl.FLOAT, false, 4*7, gl.PtrOffset(3*4))

	// 基本設定
	gl.Enable(gl.DEPTH_TEST)

	angle := 0.0
	previousTime := glfw.GetTime()

	// メインループ
	for !window.ShouldClose() {

		// 画面クリア
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.ClearColor(0.0, 0.0, 1.0, 1.0)
		gl.ClearDepth(1.0)

		// 使用するシェーダを洗濯
		gl.UseProgram(program)

		// 各行列設定
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(1280)/float32(960), 1.0, 10000.0)
		projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		camera := mgl32.LookAtV(mgl32.Vec3{0, 0, 2}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})
		cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += elapsed
		model := mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		// バッファをバインド
		gl.BindVertexArray(vao)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)

		// 描画
		gl.DrawElements(gl.TRIANGLES, int32(len(sampleModel.mMeshDataList[0].mIndices)), gl.UNSIGNED_INT, gl.PtrOffset(0))

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

// CreateShaderProgram シェーダープログラム作成関数
// 第1戻り値　プログラムインスタンス番号
// 第2戻り値　エラー内容
func CreateShaderProgram(vertexShaderSource, fragmentShaderSource string) (uint32, error) {
	// 頂点シェーダ作成
	vertexShader, err := CompileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}

	// フラグメントシェーダ（ピクセルシェーダ）作成
	fragmentShader, err := CompileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	// シェーダープログラム作成
	program := gl.CreateProgram()

	// シェーダーリンク
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to link program: %v", log)
	}

	// プログラムさえあれば動くのでシェーダー削除
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}

// CompileShader シェーダコンパイル
func CompileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csource := gl.Str(source)
	gl.ShaderSource(shader, 1, &csource, nil)
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}
