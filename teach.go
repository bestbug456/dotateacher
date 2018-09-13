package main

import (
	"fmt"
	"log"
	"math"

	rprop "github.com/bestbug456/gorpropplus"
)

func CreateNewNeuralNetworkAndValidate(args interface{}) error {

	infos := args.(*JobArgs)
	NN, err := trainNewNeuralNetwork(infos.TrainData, infos.NeuronNumber)
	if err != nil {
		return err
	}
	QAresults := checkWeightsQuality(NN, infos.TestData)
	infos.result <- NNmessage{
		MatrixQA: QAresults,
		NN:       NN,
	}
	return nil
}

func orderPickByTeamAndCreateBitmask(picks []int) []float64 {
	updatedpick := make([]int, len(picks))
	for i := 0; i < len(picks); i++ {
		updatedpick[i] = compressed[picks[i]]
	}
	team1Pick := []int{
		updatedpick[0],
		updatedpick[3],
		updatedpick[5],
		updatedpick[7],
		updatedpick[8],
	}
	team2Pick := []int{
		updatedpick[1],
		updatedpick[2],
		updatedpick[4],
		updatedpick[6],
		updatedpick[9],
	}
	bitmasks := createBitmasksForTeam(team1Pick)
	supp := createBitmasksForTeam(team2Pick)
	bitmasks = append(bitmasks, supp...)
	return bitmasks
}

func trainNewNeuralNetwork(traindata []MatchInfos, neuron int) (*rprop.NeuralNetwork, error) {
	args := rprop.NeuralNetworkArguments{
		HiddenLayer:        []int{neuron},
		InputSize:          230,
		OutputSize:         1,
		Threshold:          0.001,
		StepMax:            999999999999999999,
		LifeSignStep:       1000,
		LinearOutput:       false,
		Minus:              0.5,
		Plus:               1.2,
		ActivationFunction: rprop.Logistic,
		DerivateActivation: rprop.DerivateLogistic,
		ErrorFunction:      rprop.SSE,
		DerivateError:      rprop.DerivateSSE,
	}

	// Get a fresh new neural network
	NN, err := rprop.NewNeuralNetworkAndSetup(args)
	if err != nil {
		return nil, fmt.Errorf("Error while creating a new neural network: %s", err.Error())
	}

	inputData := make([][]float64, len(traindata))
	outputData := make([][]float64, len(traindata))
	for i := 0; i < len(traindata); i++ {
		inputData[i] = orderPickByTeamAndCreateBitmask(traindata[i].Picks)
		outputData[i] = make([]float64, 1)
		outputData[i] = []float64{float64(traindata[i].Win)}
	}

	err = NN.Train(inputData, outputData)
	if err != nil {
		return nil, fmt.Errorf("Error while training the neural network: %s", err.Error())
	}
	return NN, nil
}

func checkWeightsQuality(NN *rprop.NeuralNetwork, input []MatchInfos) []int {
	resultMatrix := make([]int, 2)
	for i := 0; i < len(input); i++ {
		testset := orderPickByTeamAndCreateBitmask(input[i].Picks)
		out, err := NN.Predict(testset)
		if err != nil {
			log.Printf("Error while predicting: %s", err.Error())
		}
		if int(out[0]) == input[i].Win && math.Abs(out[0]-float64(input[i].Win)) < 0.15 {
			resultMatrix[0]++
		} else {
			resultMatrix[1]++
		}
	}
	return resultMatrix
}

func createBitmasksForTeam(team []int) []float64 {
	bitmasks := make([]float64, 115)
	for i := 0; i < len(team); i++ {
		bitmasks[team[i]] = 1
	}
	return bitmasks
}
