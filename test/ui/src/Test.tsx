import { Add, CopyAll, Delete } from "@mui/icons-material";

import { Button, IconButton, Stack, TextField } from "@mui/material";
import { SimulationButton } from "./SimulationButton";
import { store, useTestData, useTestTimelines } from "./store";
import { DatasetCmp } from "./DatasetCmp";
import { TimelineCmp } from "./TimelineCmp";

interface TestComponentProps {
  test: string;
  // onChange: (t: TestData) => void;
  // onDelete: () => void;
  // onDuplicate: () => void;
}

export function TestComponent(props: TestComponentProps) {
  const test = useTestData(props.test);
  const timelines = useTestTimelines(props.test);

  // const handleCopyConfig = () => {
  //   navigator.clipboard.writeText(JSON.stringify(toConfig(test)));
  // };

  return (
    <Stack spacing={1}>
      <Stack direction="row" spacing={2} alignItems="center">
        <TextField
          size="small"
          label="Test name"
          value={test.name}
          sx={{ minWidth: 400 }}
          onChange={(e) => {
            store.changeTest(props.test, "name", e.target.value);
          }}
        />

        <TextField
          size="small"
          label="Number of clients"
          type="number"
          value={test.n_clients}
          onChange={(e) => {
            store.changeTest(props.test, "n_clients", parseInt(e.target.value));
          }}
        />

        <TextField
          size="small"
          label="Number of rounds"
          type="number"
          value={test.rounds}
          onChange={(e) => {
            store.changeTest(props.test, "rounds", parseInt(e.target.value));
          }}
        />

        <IconButton onClick={store.cloneTest.bind(store, props.test)}>
          <CopyAll />
        </IconButton>
        <IconButton onClick={store.removeTest.bind(store, props.test)}>
          <Delete />
        </IconButton>
        {/* <Button size="small" variant="outlined" onClick={handleCopyConfig}>
          Copy Config
        </Button> */}
        <SimulationButton config={test} />
        {/* <Button
          size="small"
          variant="outlined"
          onClick={handleDuplicateTimeline}
        >
          Duplicate Timeline
        </Button> */}
      </Stack>

      <div>
        <DatasetCmp test={props.test} />

        {timelines.map((t) => (
          <TimelineCmp
            key={t}
            test={props.test}
            timeline={t}
            n_clients={test.n_clients}
          />
        ))}

        <Button
          startIcon={<Add />}
          onClick={() => store.addTimeline(props.test)}
        >
          Add pattern
        </Button>
      </div>
    </Stack>
  );
}
