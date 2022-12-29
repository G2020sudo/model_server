/*
# Copyright (c) 2021 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	framework "tensorflow/core/framework"
	pb "tensorflow_serving"

	google_protobuf "github.com/golang/protobuf/ptypes/wrappers"
	"github.com/nfnt/resize"
	"gocv.io/x/gocv"
	"google.golang.org/grpc"
)

func run_binary_input(servingAddress string, imgPath string) {
	// Read the image in binary form
	imgBytes, err := ioutil.ReadFile(imgPath)
	if err != nil {
		log.Fatalln(err)
	}

	// Target model specification
	const MODEL_NAME string = "resnet"
	const INPUT_NAME string = "map/TensorArrayStack/TensorArrayGatherV3"
	const OUTPUT_NAME string = "softmax_tensor"

	// Create Predict Request to OVMS
	predictRequest := &pb.PredictRequest{
		ModelSpec: &pb.ModelSpec{
			Name:          MODEL_NAME,
			SignatureName: "serving_default",
			VersionChoice: &pb.ModelSpec_Version{
				Version: &google_protobuf.Int64Value{
					Value: int64(0),
				},
			},
		},
		Inputs: map[string]*framework.TensorProto{
			INPUT_NAME: &framework.TensorProto{
				Dtype: framework.DataType_DT_STRING,
				TensorShape: &framework.TensorShapeProto{
					Dim: []*framework.TensorShapeProto_Dim{
						&framework.TensorShapeProto_Dim{
							Size: int64(1),
						},
					},
				},
				StringVal: [][]byte{imgBytes},
			},
		},
	}

	// Setup connection with the model server via gRPC
	conn, err := grpc.Dial(servingAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to the grpc server: %v\n", err)
	}
	defer conn.Close()

	// Create client instance to prediction service
	client := pb.NewPredictionServiceClient(conn)

	// Send predict request and receive response
	predictResponse, err := client.Predict(context.Background(), predictRequest)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Request sent successfully")

	// Read prediction results
	responseProto := predictResponse.Outputs[OUTPUT_NAME]
	responseContent := responseProto.GetTensorContent()

	// Get details about output shape
	outputShape := responseProto.GetTensorShape()
	dim := outputShape.GetDim()
	classesNum := dim[1].GetSize()

	// Convert bytes to matrix
	outMat, err := gocv.NewMatFromBytes(1, int(classesNum), gocv.MatTypeCV32FC1, responseContent)
	outMat = outMat.Reshape(1, 1)

	// Find maximum value along with its index in the output
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(outMat)

	// Get label of the class with the highest confidence
	var label string
	if classesNum == 1000 {
		label = labels[maxLoc.X]
	} else if classesNum == 1001 {
		label = labels[maxLoc.X-1]
	} else {
		fmt.Printf("Unexpected class number in the output")
		return
	}

	fmt.Printf("Predicted class: %s\nClassification confidence: %f%%\n", label, maxVal*100)
}

func run_with_conversion(servingAddress string, imgPath string) {
	file, err := os.Open(imgPath)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer file.Close()

	// Decode file to get Image type
	decodedImg, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// Resize image to match ResNet input
	resizedImg := resize.Resize(224, 224, decodedImg, resize.Lanczos3)

	// Convert image to gocv.Mat type (HWC layout)
	imgMat, err := gocv.ImageToMatRGB(resizedImg)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	newMat := gocv.NewMat()
	// Convert type so each value is represented by float32
	// as in Mat generated by ImageToMatRGB values are represented with 8 bit precision
	imgMat.ConvertTo(&newMat, gocv.MatTypeCV32FC2)

	// Having right layout and precision, convert Mat to []bytes
	imgBytes := newMat.ToBytes()

	// Target model specification
	const MODEL_NAME string = "resnet"
	const INPUT_NAME string = "map/TensorArrayStack/TensorArrayGatherV3"
	const OUTPUT_NAME string = "softmax_tensor"

	// Create Predict Request to OVMS
	predictRequest := &pb.PredictRequest{
		ModelSpec: &pb.ModelSpec{
			Name:          MODEL_NAME,
			SignatureName: "serving_default",
			VersionChoice: &pb.ModelSpec_Version{
				Version: &google_protobuf.Int64Value{
					Value: int64(0),
				},
			},
		},
		Inputs: map[string]*framework.TensorProto{
			INPUT_NAME: &framework.TensorProto{
				Dtype: framework.DataType_DT_FLOAT,
				TensorShape: &framework.TensorShapeProto{
					Dim: []*framework.TensorShapeProto_Dim{
						&framework.TensorShapeProto_Dim{
							Size: int64(1),
						},
						&framework.TensorShapeProto_Dim{
							Size: int64(224),
						},
						&framework.TensorShapeProto_Dim{
							Size: int64(224),
						},
						&framework.TensorShapeProto_Dim{
							Size: int64(3),
						},
					},
				},
				TensorContent: imgBytes,
			},
		},
	}

	conn, err := grpc.Dial(servingAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot connect to the grpc server: %v\n", err)
	}
	defer conn.Close()

	// Create client instance to prediction service
	client := pb.NewPredictionServiceClient(conn)

	// Send predict request and receive response
	predictResponse, err := client.Predict(context.Background(), predictRequest)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Request sent successfully")

	// Read prediction results
	responseProto := predictResponse.Outputs[OUTPUT_NAME]
	responseContent := responseProto.GetTensorContent()

	// Get details about output shape
	outputShape := responseProto.GetTensorShape()
	dim := outputShape.GetDim()
	classesNum := dim[1].GetSize()

	// Convert bytes to matrix
	outMat, err := gocv.NewMatFromBytes(1, int(classesNum), gocv.MatTypeCV32FC1, responseContent)
	outMat = outMat.Reshape(1, 1)

	// Find maximum value along with its index in the output
	_, maxVal, _, maxLoc := gocv.MinMaxLoc(outMat)

	// Get label of the class with the highest confidence
	var label string
	if classesNum == 1000 {
		label = labels[maxLoc.X]
	} else if classesNum == 1001 {
		label = labels[maxLoc.X-1]
	} else {
		fmt.Printf("Unexpected class number in the output")
		return
	}

	fmt.Printf("Predicted class: %s\nClassification confidence: %f%%\n", label, maxVal*100)
}

func main() {
	servingAddress := flag.String("serving-address", "localhost:8500", "The tensorflow serving address")
	binaryInput := flag.Bool("binary-input", false, "Send JPG/PNG raw bytes")
	flag.Parse()

	if flag.NArg() > 2 {
		fmt.Println("Usage: " + os.Args[0] + " --serving-address localhost:8500 path/to/img")
		os.Exit(1)
	}

	imgPath, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	if *binaryInput {
		run_binary_input(*servingAddress, imgPath)
	} else {
		run_with_conversion(*servingAddress, imgPath)
	}
}
