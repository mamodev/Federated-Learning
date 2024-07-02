import { CopyAll, Delete, ExpandMore } from "@mui/icons-material";

import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Autocomplete,
  Button,
  IconButton,
  Slider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { TimelineCmp } from "./TimelineCmp";
import { TestData } from "./types";
import { DatasetConfig, Datasets } from "./datasets";
import { toConfig } from "./config";
import { SimulationButton } from "./SimulationButton";

interface TestComponentProps {
  test: TestData;
  onChange: (t: TestData) => void;
  onDelete: () => void;
  onDuplicate: () => void;
}

export function TestComponent(props: TestComponentProps) {
  const { test, onChange, onDelete, onDuplicate } = props;

  const handleCopyConfig = () => {
    navigator.clipboard.writeText(JSON.stringify(toConfig(test)));
  };

  return (
    <Stack spacing={1}>
      <Stack direction="row" spacing={2} alignItems="center">
        <TextField
          size="small"
          label="Test name"
          value={test.name}
          sx={{ minWidth: 400 }}
          onChange={(e) => onChange({ ...test, name: e.target.value })}
        />

        <TextField
          size="small"
          label="Number of clients"
          type="number"
          value={test.n_clients}
          onChange={(e) =>
            onChange({ ...test, n_clients: parseInt(e.target.value) })
          }
        />

        <TextField
          size="small"
          label="Number of rounds"
          type="number"
          value={test.rounds}
          onChange={(e) =>
            onChange({ ...test, rounds: parseInt(e.target.value) })
          }
        />

        <IconButton onClick={onDuplicate}>
          <CopyAll />
        </IconButton>
        <IconButton onClick={onDelete}>
          <Delete />
        </IconButton>
        <Button size="small" variant="outlined" onClick={handleCopyConfig}>
          Copy Config
        </Button>
        <SimulationButton config={test} />
      </Stack>

      <DatasetCmp
        config={test.dataset}
        onChange={(d) => onChange({ ...test, dataset: d })}
      />

      <TimelineCmp
        n_clients={test.n_clients}
        rounds={test.rounds}
        timeline={test.timeline}
        onChange={(t) => onChange({ ...test, timeline: t })}
      />
    </Stack>
  );
}

interface DatasetConfigProps {
  config: DatasetConfig;
  onChange: (c: DatasetConfig) => void;
}

export function DatasetCmp(props: DatasetConfigProps) {
  const { config, onChange } = props;

  const getClassRanges = (c: number) => {
    return Object.keys(config.indices).map(
      (k) => config.indices[parseInt(k)][c][1]
    );
  };

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
                onChange({ ...config, name: v as keyof typeof Datasets })
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
                onChange({ ...config, epochs: parseInt(e.target.value) })
              }
            />

            <TextField
              size="small"
              label="Batch size"
              type="number"
              value={config.batch_size}
              onChange={(e) =>
                onChange({ ...config, batch_size: parseInt(e.target.value) })
              }
            />

            <TextField
              size="small"
              label="Learning rate"
              type="number"
              value={config.learning_rate}
              onChange={(e) =>
                onChange({
                  ...config,
                  learning_rate: parseFloat(e.target.value),
                })
              }
            />

            <TextField
              size="small"
              label="Momentum"
              type="number"
              value={config.momentum}
              onChange={(e) =>
                onChange({ ...config, momentum: parseFloat(e.target.value) })
              }
            />
          </Stack>
          <Stack flex={1}>
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
            ))}
          </Stack>
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}

function valueLabelFormat(value: number, idx: number) {
  return `${value} C-${idx}`;
}
