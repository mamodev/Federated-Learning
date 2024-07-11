import { Add, Save } from "@mui/icons-material";
import { Button, Stack, Typography } from "@mui/material";
import { TestComponent } from "./Test";
import { Datasets } from "./datasets";
import { store, useTestList } from "./store";
import { TestData } from "./types";

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

      bias: 0.5,
      distribution_base: 0.5,
      distribution_fn: "normal",
      knolwedge_amount: 0.1,
      loading_fn: "linear",
    },
    timelines: [
      {
        name: "Default",
        timeline: {},
        rounds: 10,
      },
    ],
  };
};

function App() {
  const handleAddTest = () => {
    store.addTest(getDefaultTestData(store.getTests().length + 1));
  };

  const dowloadTestSuite = () => {
    // const data = new Blob([JSON.stringify(tests.map(toConfig))], {
    //   type: "application/json",
    // });
    // const url = URL.createObjectURL(data);
    // const a = document.createElement("a");
    // document.body.appendChild(a);
    // a.href = url;
    // a.download = "testsuite.json";
    // a.click();
    // URL.revokeObjectURL(url);
    // document.body.removeChild(a);
  };

  const testKeys = useTestList();

  return (
    <Stack p={1}>
      <Stack direction="row" spacing={2} alignItems="center">
        <Typography variant="h4">Tests</Typography>
        <Button
          size="small"
          variant="contained"
          onClick={store.persist.bind(store)}
          startIcon={<Save />}
        >
          Save
        </Button>
        <Button
          size="small"
          variant="outlined"
          onClick={store.load.bind(store)}
        >
          Load
        </Button>
        <Button size="small" variant="outlined" onClick={dowloadTestSuite}>
          Download Test Suite
        </Button>
      </Stack>
      <Stack p={2} spacing={2} justifyContent={"flex-start"}>
        {testKeys.map((test, idx) => (
          <TestComponent key={idx} test={test} />
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
