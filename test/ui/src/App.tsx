import { Add, Save } from "@mui/icons-material";
import { Button, Stack, Typography } from "@mui/material";
import { useState } from "react";
import { TestComponent } from "./Test";
import { Datasets } from "./datasets";
import { TestData } from "./types";
import { toConfig } from "./config";

const getDefaultTestData = (n: number): TestData => {
  return {
    name: `Test ${n}`,
    n_clients: 10,
    rounds: 10,
    dataset: {
      name: Datasets.MNIST.name,
      epochs: 1,
      batch_size: 256,
      learning_rate: 0.01,
      momentum: 0.9,
      indices: Datasets.MNIST.getDefaultIndices(10),
      shuffle: true,
      network: "SimpleCNN",
    },
    timeline: {},
  };
};

function App() {
  const [tests, setTests] = useState<TestData[]>([]);

  const handleAddTest = () => {
    setTests((old) => [
      ...old,
      getDefaultTestData(
        old.filter((t) => t.name.startsWith("Test")).length + 1
      ),
    ]);
  };

  const handleSave = () => {
    window.localStorage.setItem("tests", JSON.stringify(tests));
  };

  const handleLoad = () => {
    const data = window.localStorage.getItem("tests");
    if (!data) return;

    setTests(JSON.parse(data));
  };

  const handleTestChange = (idx: number, test: TestData) => {
    setTests((old) => {
      const oldData = old[idx];

      if (oldData.dataset.name !== test.dataset.name) {
        test.dataset.indices = Datasets[
          test.dataset.name as keyof typeof Datasets
        ].getDefaultIndices(test.n_clients);
      }

      if (oldData.n_clients !== test.n_clients) {
        test.dataset.indices = Datasets[
          test.dataset.name as keyof typeof Datasets
        ].getDefaultIndices(test.n_clients);
        test.timeline = {};
      }

      return old.map((t, i) => (i == idx ? test : t));
    });
  };

  const dowloadTestSuite = () => {
    const data = new Blob([JSON.stringify(tests.map(toConfig))], {
      type: "application/json",
    });
    const url = URL.createObjectURL(data);

    const a = document.createElement("a");
    document.body.appendChild(a);

    a.href = url;
    a.download = "testsuite.json";
    a.click();

    URL.revokeObjectURL(url);
    document.body.removeChild(a);
  };

  return (
    <Stack p={1}>
      <Stack direction="row" spacing={2} alignItems="center">
        <Typography variant="h4">Tests</Typography>
        <Button
          size="small"
          variant="contained"
          onClick={handleSave}
          startIcon={<Save />}
        >
          Save
        </Button>
        <Button size="small" variant="outlined" onClick={handleLoad}>
          Load
        </Button>
        <Button size="small" variant="outlined" onClick={dowloadTestSuite}>
          Download Test Suite
        </Button>
      </Stack>
      <Stack p={2} spacing={2} justifyContent={"flex-start"}>
        {tests.map((test, idx) => (
          <TestComponent
            key={idx}
            test={test}
            onChange={(t) => {
              handleTestChange(idx, t);
            }}
            onDelete={() => {
              setTests((old) => old.filter((_, i) => i != idx));
            }}
            onDuplicate={() => {
              setTests((old) => [
                ...old,
                {
                  ...JSON.parse(JSON.stringify(test)),
                  name: `${test.name} (copy)`,
                },
              ]);
            }}
          />
        ))}
        <Button
          onClick={handleAddTest}
          sx={{ maxWidth: "fit-content" }}
          startIcon={<Add />}
        >
          Add Test
        </Button>
      </Stack>
    </Stack>
  );
}

export default App;
