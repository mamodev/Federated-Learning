package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type EvalCallback func ()

func NewEvalManager(cb EvalCallback) TaskManager {
	var round int = 0
	var prevAccuracy float64 = -1

	return &SimpleTaskManager{
		GetPayloadFunc: func (client string) (Payload, error) {
			path := "./model/final.npz"
			if round == 0 {
				path = "./model/init.npz"
			}

			file, err := os.Open(path)
			if err != nil {
				return Payload{}, err
			}

			return Payload{
				MimeType: "application/octet-stream",
				Data: file,
			}, nil
		},

		RegisterResponseFunc: func (client string, response io.ReadCloser) error {
			defer response.Close()

			type EvalBody struct {
				Accuracy float64 `json:"accuracy"`
				Total int `json:"total"`
				Correct int `json:"correct"`
				Loss float64 `json:"loss"`
			}

			var body EvalBody
			err := json.NewDecoder(response).Decode(&body)
			if err != nil {
				return err
			}

			if round == 0 {
				prevAccuracy = body.Accuracy
			}

			accIncrementPrc := (body.Accuracy - prevAccuracy) / prevAccuracy * 100		
			prevAccuracy = body.Accuracy


			sign := "+"
			if accIncrementPrc < 0 {
				sign = ""
			}

			fmt.Printf("[R-%d] Acc: %.2f (%s%.2f), Loss: %.2f\n", round, body.Accuracy, sign, accIncrementPrc, body.Loss)

			if round != 0 {
				os.Remove("./model/init.npz")
				os.Rename("./model/final.npz", "./model/init.npz")
			}
			
			cb()
			round++;
			return nil
		},
	}
}