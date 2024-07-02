import os
import numpy as np
import io

class ModelsBuffer():
    def __init__(self):
        pass
    
    def push(self, model):
        pass
    
    def retrieve(self):
        pass
    
    def push_agg_model(self, model):
        pass
    
class PersistentModelsBuffer(ModelsBuffer):
    def __init__(self, folder):

        # if folder already exists, clear it
        if os.path.exists(folder):
            for file in os.listdir(folder):
                os.remove(f"{folder}/{file}")

        if not os.path.exists(folder):
            os.makedirs(folder)

        self.folder = folder
        self.models = []
        self.round = 0

    def __get_file_name(self, idx):
        return f"{self.folder}/model_{self.round}_{idx}.npz"

    def __model_to_file(self, model, idx):
        file_name = self.__get_file_name(idx)
        np.savez(file_name, **model)

    def push(self, model):  
        self.__model_to_file(model, len(self.models))
        self.models.append(model)

    def push_agg_model(self, model):
        self.__model_to_file(model, "agg")
        self.round += 1
    
    def retrieve(self):
        old_models = self.models
        self.models = []
        return old_models

class AggStrategy():
    def __init__(self):    
        pass
    def aggregate(self, models):
        raise NotImplementedError

class Aggregator():
    def __init__(self, strategy, buffer):
        self.strategy = strategy
        self.models = buffer

    def push(self, modelbuff: bytes):
        stream = io.BytesIO(modelbuff)
        model = np.load(stream)
        self.models.push(model)

    def set_default(self, modelbuff: bytes):
        stream = io.BytesIO(modelbuff)
        model = np.load(stream)
        self.models.push_agg_model(model)
    
    def aggregate(self):
        models = self.models.retrieve()
        agg = self.strategy.aggregate(models)

        self.models.push_agg_model(agg)

        stream = io.BytesIO()
        np.savez(stream, **agg)
        return stream.getvalue()
    

class AvgStrategy(AggStrategy):
    def __init__(self):
        pass

    def aggregate(self, models):
        npzFiles = models 

        avg = {}

        for key in npzFiles[0].keys():
          darr = np.array([npz[key] for npz in npzFiles])
          avg[key] = np.mean(darr, axis=0)

        stream = io.BytesIO()
        np.savez(stream, **avg)
        stream.seek(0)
        return np.load(stream)

if __name__ == '__main__':
  print("Aggregator test")
  from mnist import MNISTModel

  model = MNISTModel({
    'lr': 0.01,
    'momentum': 0.5,
    'epochs': 3,
    'batch_size': 64,
    'shuffle': False,
  })

  MODELS_FOLDER = "out"
  ROUNDS = 4

  intial_model = model.get_def_weights()
  agg = Aggregator(AvgStrategy(), PersistentModelsBuffer(MODELS_FOLDER))
  agg.set_default(intial_model)

  for _ in range(ROUNDS):
    agg.push(intial_model)
    agg.push(intial_model)
    agg.aggregate()  

  print("Aggregator test passed!")