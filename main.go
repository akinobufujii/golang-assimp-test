package main

import (
	"fmt"

	"github.com/andrebq/assimp/conv"
)

func main() {

	// シーン読み込み
	scene, err := conv.LoadAsset("Cerberus_by_Andrew_Maximov/Cerberus_LP.FBX")
	if err != nil {
		fmt.Printf("Unable to load scene.\nCause: %v", err)
	}

	fmt.Printf("%v", scene)
	// else {
	//	dumpScene(scene, *_of)
	//}
}
