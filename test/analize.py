# usage: python analize.py <path_to_folder>

import os
import sys
import json
import pandas as pd
from nn import get_model

# model naming: model_<round>_agg.npz

def analize_folder(folder) -> pd.DataFrame:
    models = os.listdir(folder)
    models = [m for m in models if m.endswith("_agg.npz")]

    if len(models) == 0:
        print(f"No models found in {folder}")
        return -1

    # models = sorted(models, key=lambda x: int(x.split("_")[1]))
    models = sorted(models, key=lambda x: int(x.split("_")[1].split(".")[0]))

    print(f"Analizing {len(models)} models in {folder}")
    print("Models:")
    for m in models:
        print(f"  {m}")

    CONFIG_FILE = folder + "/config.json"
    with open(CONFIG_FILE, "r") as f:
      c = json.load(f)

    dataset = c["dataset"]['name']
    if dataset != "MNIST":
      print(f"Dataset {dataset} not supported")
      return None
    
    net = get_model(c['dataset'])
    df = pd.DataFrame(columns=["round", "accuracy", "correct", "total"])  

    for i, m in enumerate(models):
        print(f"Analizing model {m}")
        eval = net.evaluate(open(f"{folder}/{m}", "rb").read())
        df.loc[i] = [i, eval["accuracy"], eval["correct"], eval["total"]]

    return df

def main():
    if len(sys.argv) != 2:
        print("Usage: python analize.py <path_to_folder or path_to_collection>")
        return -1

    folder = sys.argv[1]

    if not os.path.exists(folder):
        print(f"Folder {folder} does not exist")
        return -1
    
  
    # check if folder contains only subfolders
    content = os.listdir(folder)
    subfolders = [c for c in content if os.path.isdir(f"{folder}/{c}")]


    if len(subfolders) > 0:
        print(f"Folder {folder} contains subfolders, analizing all of them")
        for subfolder in subfolders:
            df = analize_folder(f"{folder}/{subfolder}")
            if df is None:
              return -1
            
            df.to_csv(f"{folder}/{subfolder}.csv", index=False)

    else:
        df = analize_folder(folder)
        if df is None:
          return -1

        df.to_csv(f"{folder}.csv", index=False)

if __name__ == "__main__":
    exit(main())