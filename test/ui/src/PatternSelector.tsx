import React from "react";

import {
  Autocomplete,
  Button,
  Dialog,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { Pattern, PatternGenerator, Patterns } from "./patterns";

type PatternComponentProps = {
  pattern: Pattern;
  onChange: (factory: PatternGenerator) => void;
  onAbort: () => void;
};

function PatternComponent(props: PatternComponentProps) {
  const [args, setArgs] = React.useState<string[]>([]);
  const [valid, setValid] = React.useState<boolean>(false);

  const handleCreate = () => {
    if (valid) {
      props.onChange(props.pattern.factory(...args.map((a) => parseFloat(a))));
    }
  };

  return (
    <Stack p={2} spacing={2}>
      <Typography>{props.pattern.name}</Typography>
      {props.pattern.args.map((arg, i) => (
        <TextField
          key={arg.name}
          type="number"
          placeholder={arg.name}
          value={args[i]}
          onChange={(e) => {
            const newArgs = [...args];

            newArgs[i] = e.target.value;
            setArgs(newArgs);

            if (newArgs.length === props.pattern.args.length) {
              setValid(true);
            }
          }}
        />
      ))}

      <Stack direction="row" spacing={2}>
        <Button onClick={handleCreate} disabled={!valid}>
          Generate pattern
        </Button>
        <Button onClick={props.onAbort}>Abort</Button>
      </Stack>
    </Stack>
  );
}

type PatternSelectorProps = {
  onApply: (factory: PatternGenerator) => void;
};

export function PatternSelector(props: PatternSelectorProps) {
  const [selectedPattern, setSelectedPattern] = React.useState<Pattern | null>(
    null
  );

  const [showDialog, setArgsDialog] = React.useState<boolean>(false);

  const handleApply = () => {
    if (selectedPattern) {
      setArgsDialog(true);
    }
  };

  return (
    <Stack direction={"row"} spacing={1}>
      <Autocomplete
        size="small"
        options={Patterns}
        getOptionLabel={(option) => option.name}
        onChange={(e, value) => setSelectedPattern(value)}
        renderInput={(params) => (
          <TextField {...params} label="Select pattern" />
        )}
        value={selectedPattern}
        sx={{
          minWidth: 200,
        }}
      />

      <Button disabled={!selectedPattern} onClick={handleApply}>
        Apply
      </Button>

      {showDialog && selectedPattern && (
        <Dialog open={true} onClose={() => setArgsDialog(false)}>
          <PatternComponent
            pattern={selectedPattern}
            onChange={(...args) => {
              setArgsDialog(false);
              props.onApply(...args);
            }}
            onAbort={() => setArgsDialog(false)}
          />
        </Dialog>
      )}
    </Stack>
  );
}
