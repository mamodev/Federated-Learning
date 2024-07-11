import { ExpandMore } from "@mui/icons-material";

import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Autocomplete,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { Datasets } from "./datasets";
import { store, useTestField } from "./store";
import { DatasetIndices } from "./DatasetIndices";

interface DatasetConfigProps {
  test: string;
}

export function DatasetCmp(props: DatasetConfigProps) {
  const config = useTestField(props.test, "dataset");

  // const getClassRanges = (c: number) => {
  //   return Object.keys(config.indices).map(
  //     (k) => config.indices[parseInt(k)][c][1]
  //   );
  // };

  return (
    <Accordion>
      <AccordionSummary expandIcon={<ExpandMore />}>
        <Typography>Dataset</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={2}>
          <Stack direction="row" spacing={2} alignItems="center">
            {/* <TextField
              disabled
              size="small"
              label="Dataset name"
              value={config.name}
              onChange={(e) => onChange({ ...config, name: e.target.value })}
            /> */}

            <Autocomplete
              sx={{ minWidth: 200 }}
              disableClearable
              size="small"
              options={Object.keys(Datasets)}
              value={config.name}
              onChange={(_, v) =>
                store.changeDatasetField(props.test, "name", v)
              }
              renderInput={(params) => (
                <TextField {...params} label="Dataset name" />
              )}
            />

            <TextField
              size="small"
              label="Epochs"
              type="number"
              value={config.epochs}
              onChange={(e) =>
                store.changeDatasetField(
                  props.test,
                  "epochs",
                  parseInt(e.target.value)
                )
              }
            />

            <TextField
              size="small"
              label="Batch size"
              type="number"
              value={config.batch_size}
              onChange={(e) =>
                store.changeDatasetField(
                  props.test,
                  "batch_size",
                  parseInt(e.target.value)
                )
              }
            />

            <TextField
              size="small"
              label="Learning rate"
              type="number"
              value={config.learning_rate}
              onChange={(e) => {
                store.changeDatasetField(
                  props.test,
                  "learning_rate",
                  parseFloat(e.target.value)
                );
              }}
            />

            <TextField
              size="small"
              label="Momentum"
              type="number"
              value={config.momentum}
              onChange={(e) => {
                store.changeDatasetField(
                  props.test,
                  "momentum",
                  parseFloat(e.target.value)
                );
              }}
            />
          </Stack>
          <Stack flex={1}>
            <DatasetIndices dataset={config} test={props.test} />
            {/* <Typography>Class ranges</Typography>
            {Datasets[config.name as keyof typeof Datasets].classes.map((c) => (
              <Stack
                key={c[0]}
                direction="row"
                justifyContent="center"
                alignItems="center"
                spacing={2}
              >
                <Typography>{c[0]}</Typography>
                <Slider
                  // orientation="vertical"
                  value={getClassRanges(c[0])}
                  // defaultValue={getClassRanges(c[0])}
                  aria-label="Temperature"
                  valueLabelDisplay="auto"
                  valueLabelFormat={valueLabelFormat}
                  disableSwap
                  min={0}
                  max={c[1] - 1}
                  track={false}
                />
              </Stack>
            ))} */}
          </Stack>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}

// function valueLabelFormat(value: number, idx: number) {
//   return `${value} C-${idx}`;
// }
