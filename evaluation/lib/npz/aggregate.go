package npz

import "fmt"

type AggFunc func ([]float64) float64

func Aggragate(dicts []NpzDict, writer *NpzDictWriter, aggregator AggFunc) error {

	if !HaveSameShape(dicts...) {
		return fmt.Errorf("npz dicts do not have the same shape")
	}

	shape := dicts[0].Arrays
	
	for key := range shape {
		w, err := writer.GetArrayWriter(key, shape[key])
		if err != nil {
			return fmt.Errorf("error getting array writer %v", err)
		}

		readers := make([]*NpyArrayReader, 0, len(dicts))
		defer func () {
			for _, reader := range readers {
				reader.Close()
			}
		}()

		for _, npz := range dicts {
			reader, err := npz.GetArrayReader(key)
			if err != nil {
				return fmt.Errorf("error getting array reader %v", err)
			}

			readers = append(readers, reader)
		}

		for range(shape[key].GetShapeSize()) {

			values := make([]float64, 0, len(readers))

			for _, reader := range readers {
				val, err := reader.Read()
				if err != nil {
					fmt.Println("Error reading value", err)
					return fmt.Errorf("error reading value %v", err)
				}

				values = append(values, val)
			}

			agg := aggregator(values)
			w.Write(agg)
		}
	}

	return nil
}

func Mean (values []float64) float64 {
	sum := 0.0
	for _, val := range values {
		sum += val
	}

	return sum / float64(len(values))
}